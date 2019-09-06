package v1

import (
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// ConfigurationChangedHandler interface specifies what method is required to support
// being a target of a ConfigurationManger object.
type ConfigurationChangedHandler interface {
	// ConfigurationChanged is a method that gets triggered when a ConfigurationManager
	// has a new configuration that should be loaded.
	ConfigurationChanged(string, config_v2.Config) error
}

// ConfigurationManagerOptions contains the configuration for the configuration
// manager instantiation.
type ConfigurationManagerOptions struct {
	// Name is the internal name that the configuraton manager will have. This
	// may be different from the name passed back from the factory.
	Name string
}

// ConfigurationManager is the interface used to obtain configuration data and
// to trigger updates
type ConfigurationManager interface {
	// Initialize is called to instantiate the ConfigurationManager and provide
	// a handler that will be notified of configuration object updates.
	Initialize(handler ConfigurationChangedHandler, configSpec string) error

	// GetName returns the internal name that the ConfigurationManager was
	// instantiated with.
	GetName() string
}
