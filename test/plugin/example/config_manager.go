package example

import (
	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

type configManager struct {
	Name string
}

// ConfigManagerFactory constructs a mock ConfigFactory
func ConfigManagerFactory(options plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager {
	return &configManager{
		Name: options.Name,
	}
}

func (manager *configManager) Initialize(handler plugin_v1.ConfigurationChangedHandler, configSpec string) error {
	configuration, err := config.LoadFromFile(configSpec)
	if err != nil {
		return err
	}

	go func() {
		handler.ConfigurationChanged(manager.Name, configuration)
	}()

	return nil
}

// GetName returns the name of the provider
func (manager *configManager) GetName() string {
	return manager.Name
}
