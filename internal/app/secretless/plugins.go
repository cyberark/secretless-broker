package secretless

import (
	"github.com/conjurinc/secretless/internal/app/secretless/listeners/http"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners/mysql"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners/pg"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners/ssh"
	"github.com/conjurinc/secretless/internal/app/secretless/listeners/sshagent"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

var InternalManagers = map[string]func() plugin_v1.ConnectionManager{}

var InternalListeners = map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
	"http":      http.ListenerFactory,
	"mysql":     mysql.ListenerFactory,
	"pg":        pg.ListenerFactory,
	"ssh":       ssh.ListenerFactory,
	"ssh-agent": sshagent.ListenerFactory,
}
