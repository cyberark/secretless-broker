package stdout

import (
	"fmt"
	"strings"
	"time"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
)

type StdoutFormatter struct{}

func NewFormatter(options formatter_api.FormatterOptions) (formatter_api.OutputFormatter, error) {
	return &StdoutFormatter{}, nil
}

func (formatter *StdoutFormatter) ProcessResults(backendNames []string, aggregatedTimings map[string]formatter_api.BackendTiming) error {
	dividerString := strings.Repeat("-", 85)
	fmt.Printf("%s\n", dividerString)

	fmt.Printf("%-20s|%15s|%15s|%15s|%8s|%8s|%13s|%15s|\n",
		"Name",
		"Min Duration",
		"Max Duration",
		"Avg Duration",
		"Runs",
		"Errors",
		"Success(%)",
		"Total Duration")

	fmt.Printf("%s\n", dividerString)

	for _, backendName := range backendNames {
		timingInfo := aggregatedTimings[backendName]

		successfulRuns := timingInfo.Count - len(timingInfo.Errors)

		averageDuration := 0 * time.Second
		if successfulRuns > 0 {
			averageDuration = time.Duration(int64(timingInfo.Duration) /
				int64(successfulRuns))
		}

		fmt.Printf("%-20s %15v %15v %15v %8d %8d %13.0f %15v \n",
			backendName,
			timingInfo.MinimumDuration,
			timingInfo.MaximumDuration,
			averageDuration,
			timingInfo.Count,
			len(timingInfo.Errors),
			(float32(successfulRuns)/float32(timingInfo.Count))*100,
			timingInfo.Duration)
	}

	return nil
}
