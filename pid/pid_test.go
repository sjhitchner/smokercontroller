package pid

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&PIDSuite{})

type PIDSuite struct{}

func (s *PIDSuite) SetUpSuite(c *C) {
	log.SetLevel(log.FatalLevel)
}

type MockInput struct {
	Value float64
}

func (t *MockInput) Read() float64 {
	return t.Value
}

type MockOutput struct {
	Value float64
}

func (t *MockOutput) Update(input, output float64) {
	t.Value = output
	fmt.Printf("temp,%0.4f\n", input)
}

type Simulator struct {
	input  *MockInput
	output *MockOutput
}

func (t *Simulator) Advance() {
	value := t.output.Value
	if value > 0.5 {
		// increase
		t.input.Value += (value - 0.5)

	} else if value < 0.5 {
		// decrease
		t.input.Value -= (value)

	} else {
		// small random change
		t.input.Value += (rand.Float64() - 0.5)

	}
}

func (s *PIDSuite) Test_1(c *C) {
	input := &MockInput{}
	output := &MockOutput{}
	simu := &Simulator{input, output}

	pid := NewProportionalBandPID(input, output, 100, 120, 60)
	pid.Setpoint = 225

	now := time.Now()
	duration := 5 * time.Second
	for i := 0; i < 1000; i++ {
		pid.NextIteration(now, duration)
		now = now.Add(duration)
		simu.Advance()
		simu.Advance()
	}
}
