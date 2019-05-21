package handlers

import (
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/http"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/ssh"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/sshagent"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// HandlerFactories contains the list of built-in handler factories
var HandlerFactories = map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
	"http":        	   http.HttpHandlerFactory,
	"mysql":           mysql.HandlerFactory,
	"ssh":             ssh.HandlerFactory,
	"sshagent":        sshagent.HandlerFactory,
}
