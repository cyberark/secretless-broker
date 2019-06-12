package timing

import (
	"fmt"
	"log"
	"time"
)

type BackendTiming struct {
	BaselineDivergencePercent map[int]int
	Count                     int
	Duration                  time.Duration
	Errors                    []TestRunError `json:"errors"`
	MaximumDuration           time.Duration
	MinimumDuration           time.Duration
}

type TestRunError struct {
	Error error `json:"error"`
	Round int   `json:"round"`
}

type SingleRunTiming struct {
	BaselineTestDuration time.Duration
	BackendName          string
	Duration             time.Duration
	MaxRounds            int
	Round                int
	TestError            error
}

type AggregateTimings struct {
	BaselineBackendName string
	Silent              bool
	Timings             map[string]BackendTiming
	processingDoneChan  chan bool
	timingReceiverChan  chan *SingleRunTiming
}

const TimingBufferScalingFactor = 100
const ZeroDuration = 0 * time.Second

func NewAggregateTimings(backendNames *[]string, baselineBackendName string,
	threads int, silent bool) AggregateTimings {

	timingChannelbufferSize := threads * len(*backendNames) * TimingBufferScalingFactor

	aggregateTimings := AggregateTimings{
		BaselineBackendName: baselineBackendName,
		Timings:             map[string]BackendTiming{},
		Silent:              silent,
		processingDoneChan:  make(chan bool),
		timingReceiverChan:  make(chan *SingleRunTiming, timingChannelbufferSize),
	}

	for _, backendName := range *backendNames {
		aggregateTimings.Timings[backendName] = BackendTiming{
			BaselineDivergencePercent: map[int]int{},
			Count:                     0,
			Duration:                  ZeroDuration,
			MinimumDuration:           ZeroDuration,
			MaximumDuration:           ZeroDuration,
			Errors:                    []TestRunError{},
		}
	}

	aggregateTimings.setupTimingReceiver()

	return aggregateTimings
}

func (aggregateTimings *AggregateTimings) AddTimingData(runTiming *SingleRunTiming) {
	aggregateTimings.timingReceiverChan <- runTiming
}

// We just wait until the timings channel is empty
func (aggregateTimings *AggregateTimings) Process() {
	close(aggregateTimings.timingReceiverChan)

	log.Println("Waiting until all the data is processed...")

	<-aggregateTimings.processingDoneChan

	log.Println("Data aggregation done!")
}

func (aggregateTimings *AggregateTimings) setupTimingReceiver() {
	go func() {
		for {
			runTiming, more := <-aggregateTimings.timingReceiverChan
			if !more {
				log.Println("Timing channel closed. Exiting...")
				aggregateTimings.processingDoneChan <- true
				return
			}

			aggregateTimings.updateBackendTiming(runTiming)
		}
	}()
}

func (aggregateTimings *AggregateTimings) updateBackendTiming(runTiming *SingleRunTiming) {
	backendTiming := aggregateTimings.Timings[runTiming.BackendName]
	backendTiming.Count++

	if runTiming.TestError != nil {
		log.Printf("[%.3d/%s] %-35s=> %v", runTiming.Round, runTiming.MaxRounds,
			runTiming.BackendName, runTiming.TestError)
		backendTiming.Errors = append(backendTiming.Errors,
			TestRunError{
				Error: runTiming.TestError,
				Round: runTiming.Round,
			})
		aggregateTimings.Timings[runTiming.BackendName] = backendTiming
		return
	}

	backendTiming.Duration = backendTiming.Duration + runTiming.Duration

	if backendTiming.MinimumDuration == ZeroDuration {
		backendTiming.MinimumDuration = backendTiming.Duration
	}

	if runTiming.Duration > backendTiming.MaximumDuration {
		backendTiming.MaximumDuration = runTiming.Duration
	}

	if runTiming.Duration < backendTiming.MinimumDuration {
		backendTiming.MinimumDuration = runTiming.Duration
	}

	baselineDivergencePercent := 100
	if runTiming.BackendName != aggregateTimings.BaselineBackendName {
		baselineDivergencePercent = int(float32(runTiming.Duration) /
			float32(runTiming.BaselineTestDuration) * 100.0)
	}

	if !aggregateTimings.Silent {
		log.Printf("[%d/%s], %-35s=>%15v, %3d%%", runTiming.Round, runTiming.MaxRounds,
			runTiming.BackendName, runTiming.Duration, baselineDivergencePercent)
	} else {
		if runTiming.BackendName == aggregateTimings.BaselineBackendName && runTiming.Round%1000 == 0 {
			fmt.Printf(".")
		}
	}

	backendTiming.BaselineDivergencePercent[baselineDivergencePercent]++

	aggregateTimings.Timings[runTiming.BackendName] = backendTiming
}
