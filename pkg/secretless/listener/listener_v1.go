package listener

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/handler"
)

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener_v1 interface {
	GetConfig() config.Listener
	GetListener() net.Listener
	GetHandlers() []handler.Handler_v1
	GetConnections() []net.Conn
}
