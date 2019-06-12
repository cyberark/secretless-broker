package stdout

import (
	"fmt"
	"strings"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/util"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
)

type StdoutFormatter struct{}

func NewFormatter(options formatter_api.FormatterOptions) (formatter_api.OutputFormatter, error) {
	return &StdoutFormatter{}, nil
}

func (formatter *StdoutFormatter) ProcessResults(backendNames []string, aggregatedTimings map[string]timing.BackendTiming, baselineThresholdMaxPercent int) error {
	fields := []map[string]string{
		map[string]string{"name": "Name", "nameFormat": "%-30s", "valueFormat": "%-30s"},
		map[string]string{"name": "Min", "nameFormat": "%13s", "valueFormat": "%13v"},
		map[string]string{"name": "Max", "nameFormat": "%13s", "valueFormat": "%13v"},
		map[string]string{"name": "Avg", "nameFormat": "%13s", "valueFormat": "%13v"},
		map[string]string{"name": "90% Lower %", "nameFormat": "%12s", "valueFormat": "%12.2f"},
		map[string]string{"name": "90% Upper %", "nameFormat": "%12s", "valueFormat": "%12.2f"},
		map[string]string{"name": "Total", "nameFormat": "%18s", "valueFormat": "%18v"},
		map[string]string{"name": "Rounds", "nameFormat": "%9s", "valueFormat": "%9d"},
		map[string]string{"name": "Errors", "nameFormat": "%9s", "valueFormat": "%9d"},
		map[string]string{"name": "Succ %", "nameFormat": "%9s", "valueFormat": "%9.2f"},
		map[string]string{"name": "Thresh %", "nameFormat": "%9s", "valueFormat": "%9.2f"},
	}

	dividerString := strings.Repeat("-", 158)

	fmt.Printf("%s\n", dividerString)
	formatValueString := ""
	for _, field := range fields {
		fmt.Printf(field["nameFormat"]+"|", field["name"])
		formatValueString += field["valueFormat"] + "|"
	}
	fmt.Printf("\n%s\n", dividerString)

	for _, backendName := range backendNames {
		timingInfo := aggregatedTimings[backendName]

		averageDuration := util.GetAverageDuration(&timingInfo)
		successPercentage := util.GetSuccessPercentage(&timingInfo)

		lowerBoundCI, upperBoundCI := util.GetConfidenceInterval90(&timingInfo.BaselineDivergencePercent)
		thresholdBreachedPercent := util.GetThresholdBreachedPercent(&timingInfo.BaselineDivergencePercent,
			baselineThresholdMaxPercent)

		fmt.Printf(formatValueString+"\n",
			backendName,
			timingInfo.MinimumDuration,
			timingInfo.MaximumDuration,
			averageDuration,
			lowerBoundCI,
			upperBoundCI,
			timingInfo.Duration,
			timingInfo.Count,
			len(timingInfo.Errors),
			successPercentage,
			thresholdBreachedPercent)
	}

	fmt.Printf("%s\n", dividerString)

	return nil
}
