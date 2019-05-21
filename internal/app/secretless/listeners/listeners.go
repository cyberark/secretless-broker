package listeners

import (
	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners/http"
	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners/mysql"
	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners/ssh"
	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners/sshagent"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// ListenerFactories contains the list of built-in listener factories
var ListenerFactories = map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
	"http":      http.ListenerFactory,
	"mysql":     mysql.ListenerFactory,
	"ssh":       ssh.ListenerFactory,
	"ssh-agent": sshagent.ListenerFactory,
}
