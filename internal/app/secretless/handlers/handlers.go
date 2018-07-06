package handlers

import (
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/http"
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/mysql"
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/pg"
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/ssh"
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/sshagent"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

// HandlerFactories contains the list of built-in handler factories
var HandlerFactories = map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
	"http/aws":    http.AWSHandlerFactory,
	"http/conjur": http.ConjurHandlerFactory,
	"mysql":       mysql.HandlerFactory,
	"pg":          pg.HandlerFactory,
	"ssh":         ssh.HandlerFactory,
	"sshagent":    sshagent.HandlerFactory,
}
