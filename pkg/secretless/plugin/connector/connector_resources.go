package connector

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

// Resources is an interface that defines what things will
// be needed by a plugin at runtime.
type Resources interface {
	// Config returns the content of addition configuration
	// parameters passed to the connector.
	Config() []byte

	// Logger returns an instance of a Logger that can be used
	// to record logging messages from the plugin.
	Logger() log.Logger
}
