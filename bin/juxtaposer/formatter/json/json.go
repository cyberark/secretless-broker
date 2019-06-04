package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
)

type JsonFormatter struct {
	Options formatter_api.FormatterOptions
}

type BackendTimingDataJson struct {
	AverageDurationNs int64                        `json:"averageDurationNs"`
	Errors            []formatter_api.TestRunError `json:"errors"`
	FailedRounds      int                          `json:"failedRounds"`
	SuccessfulRounds  int                          `json:"successfulRounds"`
	SuccessPercentage float32                      `json:"successPercentage"`
	TotalDurationNs   int64                        `json:"totalDurationNs"`
	TotalRounds       int                          `json:"totalRounds"`
}

type JsonOutput struct {
	Backends map[string]BackendTimingDataJson `json:"backends"`
}

func NewFormatter(options formatter_api.FormatterOptions) (formatter_api.OutputFormatter, error) {
	return &JsonFormatter{
		Options: options,
	}, nil
}

func (formatter *JsonFormatter) ProcessResults(backendNames []string, aggregatedTimings map[string]formatter_api.BackendTiming) error {
	jsonOutput := JsonOutput{
		Backends: map[string]BackendTimingDataJson{},
	}

	for _, backendName := range backendNames {
		timingInfo := aggregatedTimings[backendName]

		failedRounds := len(timingInfo.Errors)
		successfulRounds := timingInfo.Count - failedRounds

		averageDuration := 0 * time.Second
		if successfulRounds > 0 {
			averageDuration = time.Duration(int64(timingInfo.Duration) /
				int64(successfulRounds))
		}

		successPercentage := (float32(successfulRounds) / float32(timingInfo.Count)) * 100
		timingDataJson := BackendTimingDataJson{
			AverageDurationNs: averageDuration.Nanoseconds(),
			Errors:            timingInfo.Errors,
			FailedRounds:      failedRounds,
			SuccessfulRounds:  successfulRounds,
			SuccessPercentage: successPercentage,
			TotalDurationNs:   timingInfo.Duration.Nanoseconds(),
			TotalRounds:       timingInfo.Count,
		}

		jsonOutput.Backends[backendName] = timingDataJson
	}

	timingInfoBytes, err := json.MarshalIndent(jsonOutput, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	outputFilename := formatter.Options["output_file"]
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
