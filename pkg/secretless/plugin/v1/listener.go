package v1

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// ListenerOptions contains the configuration for the listener
type ListenerOptions struct {
	EventNotifier  EventNotifier
	HandlerConfigs []config.Handler
	ListenerConfig config.Listener
	NetListener    net.Listener
	Resolver       Resolver
	RunHandlerFunc func(string, HandlerOptions) Handler
}

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener interface {
	GetConfig() config.Listener
	GetConnections() []net.Conn
	GetHandlers() []Handler
	GetListener() net.Listener
	GetName() string
	GetNotifier() EventNotifier
	Listen()
	Validate() error
	Shutdown() error
}
