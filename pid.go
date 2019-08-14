package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	ppid "github.com/sjhitchner/smokercontroller/pid"
)

var (
	TemperatureBand float64
	SetPoint        float64
	IntegrationTime float64
	DerivativeTime  float64
	SampleTime      time.Duration
)

func init() {
	log.SetLevel(log.DebugLevel)

	flag.Float64Var(&TemperatureBand, "tb", 100, "Temperature band to indicate response")
	flag.Float64Var(&SetPoint, "sp", 225, "Temperature set point (goal)")
	flag.Float64Var(&IntegrationTime, "it", 180, "Integration time to remove past error in seconds")
	flag.Float64Var(&DerivativeTime, "dt", 45, "Derivative time to remove future error in seconds")
	flag.DurationVar(&SampleTime, "t", 5*time.Second, "Sample / Cylce Time")
}

type Thermometer struct {
	Value float64
}

func (t *Thermometer) Read() float64 {
	fmt.Printf("Thermometer=%0.4f\n", t.Value)
	return t.Value
}

/*
func NewSmokerPID(input ppid.Input, output ppid.Output) *ppid.PID {
	pid := ppid.NewPID(input, output)
	pid.Kp = Kp
	pid.Ki = Ki
	pid.Kd = Kd
	return pid
}
*/

const (
	IgniterOnTemp = 100
)

type Traeger struct {
	Auger   float64
	Fan     float64
	Igniter bool
}

func (t *Traeger) Update(input, output float64) {
	t.Igniter = input < IgniterOnTemp

	t.Auger = output
	t.Fan = output

	log.WithFields(log.Fields{
		"auger":   t.Auger,
		"fan":     t.Fan,
		"igniter": t.Igniter,
	}).Infof("Trager Ctrl")

}

func main() {
	flag.Parse()

	therm := &Thermometer{200}
	traeger := &Traeger{}

	//pid := NewSmokerPID(therm, motor)
	pid := ppid.NewProportionalBandPID(therm, traeger, TemperatureBand, IntegrationTime, DerivativeTime)
	defer pid.Stop()

	pid.Setpoint = 225
	go pid.Start(5 * time.Second)

	for {
		value := traeger.Auger
		if value > 0.5 {
			// increase
			therm.Value += (value - 0.5)

		} else if value < 0.5 {
			// decrease
			therm.Value -= (value)

		} else {
			// small random change
			therm.Value += (rand.Float64() - 0.5)

		}
		<-time.After(1 * time.Second)
	}
}

func ReadTemperature(therm *Thermometer) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Enter Temperature:")
		for scanner.Scan() {
			text := scanner.Text()
			temperature, err := strconv.ParseFloat(text, 64)
			if err != nil {
				log.Errorf("Invalid temperature '%s'", text)
				break
			}
			fmt.Printf("Setting Temperature: %f\n", temperature)
			therm.Value = temperature
		}

		if scanner.Err() != nil {
			// handle error.
		}
	}
}
