package handlers

import (
	"github.com/cyberark/secretless-broker/internal/handlers/http"
	"github.com/cyberark/secretless-broker/internal/handlers/sshagent"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// HandlerFactories contains the list of built-in handler factories
var HandlerFactories = map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
	"http/basic_auth": http.BasicAuthHandlerFactory,
	"http/conjur":     http.ConjurHandlerFactory,
	"sshagent":        sshagent.HandlerFactory,
}
