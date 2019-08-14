package pid

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

type Input interface {
	Read() float64
}

type Output interface {
	Update(input float64, output float64)
}

/*
http://engineeredmusings.com/pismoker/
https://www.teachmemicro.com/arduino-pid-control-tutorial/

proportional
- motor current is set in proportion to existing error

integral
- increases action in relation not only to the error but also the time for which it has persisted

differential
- does not consider the error (meaning it cannot bring it to zero: a pure D controller cannot bring
  the system to its setpoint), but the rate of change of error, trying to bring this rate to zero.
*/

/*
#PID controller based on proportional band in standard PID form https://en.wikipedia.org/wiki/PID_controller#Ideal_versus_standard_PID_form
# u = Kp (e(t)+ 1/Ti INT + Td de/dt)
# PB = Proportional Band
# Ti = Goal of eliminating in Ti seconds
# Td = Predicts error value at Td in seconds


Parameters = {'mode': 'Off', 'target':225, 'PB': 60.0, 'Ti': 180.0, 'Td': 45.0, 'CycleTime': 20, 'u': 0.15, 'PMode': 2.0, 'program': False, 'ProgramToggle': time.time()}  #60,180,45 held +- 5F

def CalculateGains(self,PB,Ti,Td):
		self.Kp = -1/PB
		self.Ki = self.Kp/Ti
		self.Kd = self.Kp*Td
		logger.info('PB: %f Ti: %f Td: %f --> Kp: %f Ki: %f Kd: %f',PB,Ti,Td,self.Kp,self.Ki,self.Kd)

self.Inter_max = abs(0.5/self.Ki)

*/

type PID struct {
	Setpoint float64

	Kp    float64
	Ki    float64
	Kd    float64
	KiMax float64

	previousError  float64
	cumlativeError float64
	previousTime   time.Time

	input  Input
	output Output
	end    chan bool
}

// Proportional Action
// MV(t) = Kp (-PV(t) + 1/Ti int(e(t),0, t) - Td  PV(t) d/dt

func NewPID(input Input, output Output) *PID {
	return &PID{
		input:  input,
		output: output,
	}
}

// Proportional Band PID
//
// input: input for process/measured value
// output: handle output
// pb: proportional band, temperature band centered on the set point
//     if err > PB/2 output is 0
//        err < PB/2 output is 1
// ti: integral time in seconds, goal to eliminate pass error
// td: derivative time in seconds, goal to eliminate future error
//
func NewProportionalBandPID(input Input, output Output, pb, ti, td float64) *PID {
	pid := NewPID(input, output)

	pid.Kp = 1 / pb
	pid.Ki = pid.Kp / ti
	pid.Kd = pid.Kp * td
	pid.KiMax = math.Abs(0.5 / pid.Ki)
	return pid
}

// Standard Form
//
// u(t) = Kp e(t) + Ki (e(t')dt')| + Kd de(t)/dt
// u(t) = Kp ( e(t) + 1/Ti (e(t')dt')| + Td de(t)/dt)
//
// Ki = Kp/Ti
// Kd = Kp*Td
//
// Ti: Integral time, eliminating past error in Ti seconds
// Td: Derivative time, predict future error in Td seconds
//
func (t *PID) Start(duration time.Duration) {
	log.WithFields(log.Fields{
		"Kp":    t.Kp,
		"Ki":    t.Ki,
		"Kd":    t.Kd,
		"KiMax": t.KiMax,
	}).Infof("Starting PID")

	tick := time.Tick(duration)
	for {
		select {
		case <-tick:
			t.NextIteration(time.Now(), duration)

		case <-t.end:
			log.Infof("Stopping PID")
			return
		}
	}
}

// NextIteration
// Not normally used only for testing
func (t *PID) NextIteration(now time.Time, increment time.Duration) {
	elapsedTime := float64(now.Sub(t.previousTime) / time.Second)

	input := t.input.Read()

	currentError := t.Setpoint - input

	log.WithFields(log.Fields{
		"measurement": input,
		"setpoint":    t.Setpoint,
		"error":       currentError,
		"dt":          elapsedTime,
	}).Infof("Input")

	// Proportional
	proportional := math.Max(0, math.Min(1, t.Kp*currentError+0.5))

	// Integral
	t.cumlativeError = t.cumlativeError + currentError*elapsedTime
	t.cumlativeError = math.Max(t.cumlativeError, -t.KiMax)
	t.cumlativeError = math.Min(t.cumlativeError, t.KiMax)
	integral := t.Ki * t.cumlativeError

	// Derivative
	rateError := (currentError - t.previousError) / elapsedTime
	derivative := t.Kd * rateError

	output := proportional + integral + derivative

	log.WithFields(log.Fields{
		"proportional": proportional,
		"integral":     integral,
		"derivative":   derivative,
		"output":       output,
	}).Infof("Output")

	t.previousError = currentError
	t.previousTime = now

	t.output.Update(input, output)
}

func (t *PID) Stop() {
	t.end <- true
}
