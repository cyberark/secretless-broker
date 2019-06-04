package api

import (
	"time"
)

type BackendTiming struct {
	Count    int
	Duration time.Duration
	Errors   []TestRunError `json:"errors"`
}

type OutputFormatter interface {
	ProcessResults([]string, map[string]BackendTiming) error
}

type TestRunError struct {
	Error error `json:"error"`
	Round int   `json:"round"`
}

type FormatterOptions map[string]string
type FormatterConstructor func(FormatterOptions) (OutputFormatter, error)
