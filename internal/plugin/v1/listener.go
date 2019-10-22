package v1

import (
	"net"

	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// ListenerOptions contains thetype Proxy struct { configuration for the listener
type ListenerOptions struct {
	EventNotifier  EventNotifier
	ServiceConfig  config_v2.Service
	NetListener    net.Listener
	Resolver       Resolver
	RunHandlerFunc func(string, HandlerOptions) Handler
}

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener interface {
	GetConfig() config_v2.Service
	GetConnections() []net.Conn
	GetHandlers() []Handler
	GetListener() net.Listener
	GetName() string
	GetNotifier() EventNotifier
	Listen()
	Validate() error
	Shutdown() error
}
