package api

import (
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

type OutputFormatter interface {
	ProcessResults([]string, map[string]BackendTiming, int) error
}

type TestRunError struct {
	Error error `json:"error"`
	Round int   `json:"round"`
}

type FormatterOptions map[string]string
type FormatterConstructor func(FormatterOptions) (OutputFormatter, error)

func NewBackendTiming() BackendTiming {
	return BackendTiming{
		BaselineDivergencePercent: map[int]int{},
		Count:                     0,
		Duration:                  0 * time.Second,
		MinimumDuration:           0 * time.Second,
		MaximumDuration:           0 * time.Second,
		Errors:                    []TestRunError{},
	}
}
