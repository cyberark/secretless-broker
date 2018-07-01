package plugin_v1

import (
	"net"
	"net/http"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/pkg/secretless/config"
)

type HandlerOptions struct {
	HandlerConfig    config.Handler
	Channels         <-chan ssh.NewChannel
	ClientConnection net.Conn
	EventNotifier    EventNotifier
}

// Handler is an interface which takes a connection and connects it to a backend
// TODO: Remove Authenticate as it's only used by http listener
// TODO: Remove LoadKeys as it's only used by sshagent listener
type Handler interface {
	Authenticate(map[string][]byte, *http.Request) error
	GetConfig() config.Handler
	GetClientConnection() net.Conn
	GetBackendConnection() net.Conn
	LoadKeys(keyring agent.Agent) error
}
