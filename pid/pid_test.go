package pid

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&PIDSuite{})

type PIDSuite struct {
}

func (s *PIDSuite) SetUpSuite(c *C) {
	log.SetLevel(log.FatalLevel)
}

func (s *PIDSuite) TearDownSuite(c *C) {

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

func (s *PIDSuite) Test_RunSimulation1(c *C) {
	series := []chart.Series{}

	//series = append(series, RunSimulation(50, 60, 45))
	//series = append(series, RunSimulation(50, 60, 60))
	//series = append(series, RunSimulation(50, 60, 90))
	//series = append(series, RunSimulation(50, 120, 45))
	//series = append(series, RunSimulation(50, 120, 60))
	series = append(series, RunSimulation(60, 120, 90))
	//series = append(series, RunSimulation(50, 180, 45))
	//series = append(series, RunSimulation(50, 180, 60))
	//series = append(series, RunSimulation(50, 180, 90))

	PlotWithName("simulation-50.png", "PB=50", series)
}

func (s *PIDSuite) Test_RunSimulation2(c *C) {
	series := []chart.Series{}
	series = append(series, RunSimulation(100, 60, 45))
	series = append(series, RunSimulation(100, 60, 60))
	series = append(series, RunSimulation(100, 60, 90))
	series = append(series, RunSimulation(100, 120, 45))
	series = append(series, RunSimulation(100, 120, 60))
	series = append(series, RunSimulation(100, 120, 90))
	series = append(series, RunSimulation(100, 180, 45))
	series = append(series, RunSimulation(100, 180, 60))
	series = append(series, RunSimulation(100, 180, 90))

	PlotWithName("simulation-100.png", "PB=100", series)
}

func (s *PIDSuite) Test_RunSimulation3(c *C) {
	series := []chart.Series{}
	series = append(series, RunSimulation(150, 60, 45))
	series = append(series, RunSimulation(150, 60, 60))
	series = append(series, RunSimulation(150, 60, 90))
	series = append(series, RunSimulation(150, 120, 45))
	series = append(series, RunSimulation(150, 120, 60))
	series = append(series, RunSimulation(150, 120, 90))
	series = append(series, RunSimulation(150, 180, 45))
	series = append(series, RunSimulation(150, 180, 60))
	series = append(series, RunSimulation(150, 180, 90))

	PlotWithName("simulation-150.png", "PB=150", series)
}

func RunSimulation(pb, ti, td float64) chart.ContinuousSeries {
	input := &MockInput{}
	output := &MockOutput{}
	simu := &Simulator{input, output}

	pid := NewProportionalBandPID(input, output, pb, ti, td)
	pid.Setpoint = 225

	ts := chart.ContinuousSeries{
		Name: fmt.Sprintf("pb=%0.2f, ti=%0.2f, td=%0.2f", pb, ti, td),
		Style: chart.Style{
			Show: true,
		},
	}

	now := time.Now()
	duration := 5 * time.Second
	for i := 0; i < 1000; i++ {
		pid.NextIteration(now, duration)
		now = now.Add(duration)
		simu.Advance()
		simu.Advance()

		ts.XValues = append(ts.XValues, float64(i*5))
		ts.YValues = append(ts.YValues, input.Value)
	}

	return ts
}

func PlotWithName(filename, title string, series []chart.Series) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return Plot(f, title, series)
}

func Plot(w io.Writer, title string, series []chart.Series) error {
	graph := chart.Chart{
		Title: title,
		TitleStyle: chart.Style{
			Show: true,
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
			AxisType: chart.YAxisPrimary,
		},
		Series: series,
	}
	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph, chart.Style{
			Padding: chart.NewBox(100, 100, 100, 100),
		}),
	}
	return graph.Render(chart.PNG, w)
}
