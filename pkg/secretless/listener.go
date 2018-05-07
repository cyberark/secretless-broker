package secretless

import (
	"net"

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
