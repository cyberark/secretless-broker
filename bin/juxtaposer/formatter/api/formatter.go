package api

import (
	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
)

type OutputFormatter interface {
	ProcessResults([]string, map[string]timing.BackendTiming, int) error
}

type FormatterOptions map[string]string
type FormatterConstructor func(FormatterOptions) (OutputFormatter, error)
