package util

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

// Accept listeners for new connections from Listener `l` and notifies plugins
// of new connections
func Accept(l plugin_v1.Listener) (net.Conn, error) {
	conn, err := l.GetListener().Accept()
	if conn != nil && err == nil {
		l.GetNotifier().NewConnection(l, conn)
	}
	return conn, err
}
