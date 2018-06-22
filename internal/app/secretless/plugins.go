package secretless

import (
	"github.com/conjurinc/secretless/internal/app/secretless/handlers"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

var InternalHandlers = handlers.HandlerFactories
var InternalListeners = listeners.ListenerFactories
var InternalManagers = map[string]func() plugin_v1.ConnectionManager{}
