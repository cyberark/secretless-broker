package util

import (
	"net"

	plugin_v1 "github.com/conjurinc/secretless-broker/pkg/secretless/plugin/v1"
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
