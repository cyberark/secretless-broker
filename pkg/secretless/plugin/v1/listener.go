package v1

import (
	"net"
	"log"
	"sync"

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

type BaseListener struct {
	self		   Listener
	handlers       []Handler
	EventNotifier  EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	Resolver       Resolver
	Config         config.Listener
	RunHandlerFunc func(id string, options HandlerOptions) Handler
}

func NewBaseListener(options ListenerOptions, self Listener) BaseListener {
	return BaseListener{
		self:           self,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		Config:         options.ListenerConfig,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}

// GetConfig implements plugin_v1.Listener
func (l *BaseListener) GetConfig() config.Listener {
	return l.Config
}

// GetConnections implements plugin_v1.Listener
func (l *BaseListener) GetConnections() []net.Conn {
	return nil
}

// GetHandlers implements plugin_v1.Listener
func (l *BaseListener) GetHandlers() []Handler {
	return l.handlers
}

// GetListener implements plugin_v1.Listener
func (l *BaseListener) GetListener() net.Listener {
	return l.NetListener
}

func (l *BaseListener) GetName() string {
	panic("BaseListener does not implement GetName")
}

// GetNotifier implements plugin_v1.Listener
func (l *BaseListener) GetNotifier() EventNotifier {
	return l.EventNotifier
}

// Listen implements plugin_v1.Listener
func (l *BaseListener) Listen() {
	panic("BaseListener does not implement Listen")
}

// Validate implements plugin_v1.Listener
func (l *BaseListener) Validate() error {
	panic("BaseListener does not implement Validate")
}

// Shutdown implements plugin_v1.Listener
func (l *BaseListener) Shutdown() error {
	// TODO: Clean up all handlers
	self := l.self

	log.Printf("Shutting down '%v' listener", self.GetName())

	log.Printf("Shutting down '%v' listener's handlers...", self.GetName())
	var wg sync.WaitGroup

	for _, handler := range self.GetHandlers() {
		wg.Add(1)

		go func(h Handler) {
			defer wg.Done()
			h.Shutdown()
		}(handler)
	}

	wg.Wait()

	return l.NetListener.Close()
}
