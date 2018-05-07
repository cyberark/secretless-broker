package secretless

import (
	"net"

	"github.com/conjurinc/secretless/internal/pkg/plugin"
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener interface {
	GetConfig() config.Listener
	GetListener() net.Listener
	GetHandlers() []Handler
	GetConnections() []net.Conn
}

// Accept listeners for new connections from Listener `l` and notifies plugins
// of new connections
func Accept(l Listener) (net.Conn, error) {
	conn, err := l.GetListener().Accept()
	if conn != nill && err == nil {
		plugin.GetManager().NewConnection(l, conn)
	}
	return conn, err
}
