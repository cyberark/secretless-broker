package timing

import (
	"fmt"
	"log"
	"strconv"
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
	Round                int
	TestError            error
	Thread               int
}

type AggregateTimings struct {
	BaselineBackendName         string
	BaselineMaxThresholdPercent int
	MaxRoundsString             string
	Silent                      bool
	Timings                     map[string]BackendTiming
	processingDoneChan          chan bool
	timingReceiverChan          chan *SingleRunTiming
}

type AggregateTimingOptions struct {
	BackendNames                []string
	BaselineBackendName         string
	BaselineMaxThresholdPercent int
	MaxRounds                   int
	Threads                     int
	Silent                      bool
}

const TimingBufferScalingFactor = 100
const ZeroDuration = 0 * time.Second

func NewAggregateTimings(options *AggregateTimingOptions) AggregateTimings {
	timingChannelbufferSize := options.Threads * len(options.BackendNames) * TimingBufferScalingFactor

	maxRounds := "infinity"
	if options.MaxRounds >= 0 {
		maxRounds = strconv.Itoa(options.MaxRounds)
	}

	aggregateTimings := AggregateTimings{
		BaselineBackendName:         options.BaselineBackendName,
		BaselineMaxThresholdPercent: options.BaselineMaxThresholdPercent,
		MaxRoundsString:             maxRounds,
		Timings:                     map[string]BackendTiming{},
		Silent:                      options.Silent,
		processingDoneChan:          make(chan bool),
		timingReceiverChan:          make(chan *SingleRunTiming, timingChannelbufferSize),
	}

	for _, backendName := range options.BackendNames {
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

	log.Println("async: Waiting until all the data is processed...")

	<-aggregateTimings.processingDoneChan

	log.Println("Data aggregation done!")
}

func (aggregateTimings *AggregateTimings) setupTimingReceiver() {
	go func() {
		for runTiming := range aggregateTimings.timingReceiverChan {
			aggregateTimings.updateBackendTiming(runTiming)
		}

		aggregateTimings.processingDoneChan <- true
	}()
}

func (aggregateTimings *AggregateTimings) updateBackendTiming(runTiming *SingleRunTiming) {
	backendTiming := aggregateTimings.Timings[runTiming.BackendName]
	backendTiming.Count++

	backendPrintableName := fmt.Sprintf("%s/%d", runTiming.BackendName, runTiming.Thread)
	if runTiming.TestError != nil {
		log.Printf("[%.3d/%s] %-35s=> %v", runTiming.Round, aggregateTimings.MaxRoundsString,
			backendPrintableName, runTiming.TestError)
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
		log.Printf("[%d/%s], %-35s=>%15v, %4d%%", runTiming.Round,
			aggregateTimings.MaxRoundsString, backendPrintableName,
			runTiming.Duration, baselineDivergencePercent)
	} else {
		if runTiming.BackendName == aggregateTimings.BaselineBackendName && runTiming.Round%1000 == 0 {
			fmt.Printf(".")
		}
	}

	backendTiming.BaselineDivergencePercent[baselineDivergencePercent]++

	aggregateTimings.Timings[runTiming.BackendName] = backendTiming
}
