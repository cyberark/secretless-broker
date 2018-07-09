package secretless

import (
	"github.com/conjurinc/secretless/internal/app/secretless/handlers"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// InternalHandlers is the set of built-in project handlers
var InternalHandlers = handlers.HandlerFactories

// InternalListeners is the set of built-in project listeners
var InternalListeners = listeners.ListenerFactories

// InternalConnectionManagers is empty; there are currently no built-in project managers
var InternalConnectionManagers = map[string]func() plugin_v1.ConnectionManager{}
