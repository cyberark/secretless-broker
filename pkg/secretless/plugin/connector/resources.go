package connector

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

/*
Resources is an interface that exposes everything your Connector needs from
the Secretless framework and is passed to your plugin's
constructor.  You need to retain a reference to connector.Resources, via
a closure, inside the Connector function returned by your constructor.
*/
type Resources interface {
	/*
		Config() provides your connector with resources specified in your
		secretless.yml file.

		Anything specified in your connector's `config` section of the
		secretless.yml is passed back via this method as a raw []byte. Not all
		connectors require data to be passed back in this field. Your code is
		responsible for casting the raw config bytes back into a meaningful
		`struct` that your code can work with.
	*/
	Config() []byte

	/*
		Logger() provides a basic logger you can use for debugging and
		informational logging.

		This method returns an object similar to the standard library's
		`log.Logger` to let you log events to stdout and stderr.  It respects
		Secretless's command line `debug` flag, so that calling Debugf or Infof
		does nothing unless you started Secretless in debug mode.
	*/
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
// to be used in invocation of a new Connector.
func NewResources(
	config []byte,
	logger log.Logger) Resources {

	return &_resources{
		config: config,
		logger: logger,
	}
}

// CredentialValuesByID is a type that maps credential IDs to their values.
type CredentialValuesByID map[string][]byte
