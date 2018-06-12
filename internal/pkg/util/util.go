package util

import (
	"net"

	"github.com/conjurinc/secretless/internal/pkg/plugin"
	"github.com/conjurinc/secretless/pkg/secretless/listener"
)

// Accept listeners for new connections from Listener `l` and notifies plugins
// of new connections
func Accept(l listener.Listener_v1) (net.Conn, error) {
	conn, err := l.GetListener().Accept()
	if conn != nil && err == nil {
		plugin.GetManager().NewConnection(l, conn)
	}
	return conn, err
}
