package plugin_v1

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless/config"
)

type ListenerOptions struct {
	EventNotifier  EventNotifier
	HandlerConfigs []config.Handler
	ListenerConfig config.Listener
	NetListener    net.Listener
	RunHandlerFunc func(string, HandlerOptions) Handler
}

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener interface {
	Validate() error
	GetConfig() config.Listener
	GetNotifier() EventNotifier
	GetListener() net.Listener
	GetHandlers() []Handler
	GetConnections() []net.Conn
	Listen()
}
