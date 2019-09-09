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

type _resources struct {
	config []byte
	logger log.Logger
}

func (res *_resources) Config() []byte {
	return res.config
}

func (res *_resources) Logger() log.Logger {
	return res.logger
}

// NewResources creates a new Resources interface from a backing object struct
// to be used in invocation of a new Connector
func NewResources(
	config []byte,
	logger log.Logger) Resources {

	return &_resources{
		config: config,
		logger: logger,
	}
}
