package internal

import (
	"github.com/cyberark/secretless-broker/internal/configurationmanagers"
	"github.com/cyberark/secretless-broker/internal/handlers"
	"github.com/cyberark/secretless-broker/internal/listeners"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

// InternalHandlers is the set of built-in handlers
var InternalHandlers = handlers.HandlerFactories

// InternalListeners is the set of built-in listeners
var InternalListeners = listeners.ListenerFactories

// InternalProviders is the set of built-in providers
var InternalProviders = providers.ProviderFactories

// InternalConfigurationManagers is the set of built-in configuration managers
var InternalConfigurationManagers = configurationmanagers.ConfigurationManagerFactories

// InternalConnectionManagers is empty; there are currently no built-in project managers
var InternalConnectionManagers = map[string]func() plugin_v1.ConnectionManager{}
