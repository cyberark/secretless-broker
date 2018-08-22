package secretless

import (
	"github.com/cyberark/secretless-broker/internal/app/secretless/configuration_managers"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers"
	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners"
	"github.com/cyberark/secretless-broker/internal/app/secretless/providers"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// InternalHandlers is the set of built-in handlers
var InternalHandlers = handlers.HandlerFactories

// InternalListeners is the set of built-in listeners
var InternalListeners = listeners.ListenerFactories

// InternalProviders is the set of built-in providers
var InternalProviders = providers.ProviderFactories

// InternalConfigurationManagers is the set of built-in configuration managers
var InternalConfigurationManagers = configuration_managers.ConfigurationManagerFactories

// InternalConnectionManagers is empty; there are currently no built-in project managers
var InternalConnectionManagers = map[string]func() plugin_v1.ConnectionManager{}
