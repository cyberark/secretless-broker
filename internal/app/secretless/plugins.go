package secretless

import (
	"github.com/conjurinc/secretless/internal/app/secretless/http"
	"github.com/conjurinc/secretless/internal/app/secretless/mysql"
	"github.com/conjurinc/secretless/internal/app/secretless/pg"
	"github.com/conjurinc/secretless/internal/app/secretless/ssh"
	"github.com/conjurinc/secretless/internal/app/secretless/sshagent"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

var InternalListeners = map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
	"http":      http.ListenerFactory,
	"mysql":     mysql.ListenerFactory,
	"pg":        pg.ListenerFactory,
	"ssh":       ssh.ListenerFactory,
	"ssh-agent": sshagent.ListenerFactory,
}
