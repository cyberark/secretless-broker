package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/util"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
)

type JsonFormatter struct {
	Options formatter_api.FormatterOptions
}

type ConfidenceIntervalJson struct {
	IntervalPercent   int     `json:"intervalPercent"`
	LowerBoundPercent float64 `json:"lowerBoundPercent"`
	UpperBoundPercent float64 `json:"upperBoundPercent"`
}

type BackendTimingDataJson struct {
	AverageDurationNs        int64                  `json:"averageDurationNs"`
	ConfidenceInterval       ConfidenceIntervalJson `json:"confidenceInterval"`
	Errors                   []timing.TestRunError  `json:"errors"`
	FailedRounds             int                    `json:"failedRounds"`
	MaximumDurationNs        int64                  `json:"maximumDurationNs"`
	MinimumDurationNs        int64                  `json:"minimumDurationNs"`
	SuccessfulRounds         int                    `json:"successfulRounds"`
	SuccessPercentage        float64                `json:"successPercentage"`
	ThresholdBreachedPercent float64                `json:"thresholdBreachedPercent"`
	TotalDurationNs          int64                  `json:"totalDurationNs"`
	TotalRounds              int                    `json:"totalRounds"`
}

type JsonOutput struct {
	Backends map[string]BackendTimingDataJson `json:"backends"`
}

func NewFormatter(options formatter_api.FormatterOptions) (formatter_api.OutputFormatter, error) {
	return &JsonFormatter{
		Options: options,
	}, nil
}

func (formatter *JsonFormatter) ProcessResults(backendNames []string,
	aggregatedTimings map[string]timing.BackendTiming, baselineThresholdMaxPercent int) error {

	jsonOutput := JsonOutput{
		Backends: map[string]BackendTimingDataJson{},
	}

	for _, backendName := range backendNames {
		timingInfo := aggregatedTimings[backendName]

		failedRounds := len(timingInfo.Errors)
		successfulRounds := timingInfo.Count - failedRounds

		averageDuration := util.GetAverageDuration(&timingInfo)
		successPercentage := util.GetSuccessPercentage(&timingInfo)

		lowerBoundCI, upperBoundCI := util.GetConfidenceInterval90(&timingInfo.BaselineDivergencePercent)
		thresholdBreachedPercent := util.GetThresholdBreachedPercent(&timingInfo.BaselineDivergencePercent,
			baselineThresholdMaxPercent)

		timingDataJson := BackendTimingDataJson{
			AverageDurationNs: averageDuration.Nanoseconds(),
			Errors:            timingInfo.Errors,
			FailedRounds:      failedRounds,
			ConfidenceInterval: ConfidenceIntervalJson{
				IntervalPercent:   90,
				LowerBoundPercent: lowerBoundCI,
				UpperBoundPercent: upperBoundCI,
			},
			MaximumDurationNs:        timingInfo.MaximumDuration.Nanoseconds(),
			MinimumDurationNs:        timingInfo.MinimumDuration.Nanoseconds(),
			SuccessfulRounds:         successfulRounds,
			SuccessPercentage:        successPercentage,
			ThresholdBreachedPercent: thresholdBreachedPercent,
			TotalDurationNs:          timingInfo.Duration.Nanoseconds(),
			TotalRounds:              timingInfo.Count,
		}

		jsonOutput.Backends[backendName] = timingDataJson
	}

	timingInfoBytes, err := json.MarshalIndent(jsonOutput, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	outputFilename := formatter.Options["outputFile"]
	if outputFilename == "" {
		fmt.Printf("%s\n", timingInfoBytes)
		return nil
	}

	err = ioutil.WriteFile(outputFilename, timingInfoBytes, 0644)
	if err == nil {
		log.Printf("Successfully wrote JSON results to file '%s'.", outputFilename)
	}

	return err
}
