package v1

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"log"
	"net"
	"sync"
)

// ListenerOptions contains thetype Proxy struct { configuration for the listener
type ListenerOptions struct {
	EventNotifier  EventNotifier
	HandlerConfigs []v1.Handler
	ListenerConfig v1.Listener
	NetListener    net.Listener
	Resolver       Resolver
	RunHandlerFunc func(string, HandlerOptions) Handler
}

// Listener is the interface which accepts client connections and passes them
// to a handler
type Listener interface {
	GetConfig() v1.Listener
	GetConnections() []net.Conn
	GetHandlers() []Handler
	GetListener() net.Listener
	GetName() string
	GetNotifier() EventNotifier
	Listen()
	Validate() error
	Shutdown() error
}

// BaseListener provides default (shared/common) implementations
// of Listener interface methods, where it makes sense
// - the rest of the methods panic if
// not implemented in the "DerivedListener"
// e.g. BaseListener#GetName.
//
// The intention is to keep things DRY by
// embedding BaseListener in "DerivedListener".
//
// There is no requirement to use BaseListener.
type BaseListener struct {
	closingMutex   *sync.Mutex
	handlers       []Handler // store of active handlers for this listener,
	Config         v1.Listener
	EventNotifier  EventNotifier
	HandlerConfigs []v1.Handler
	IsClosed       bool
	NetListener    net.Listener
	Resolver       Resolver
	RunHandlerFunc func(id string, options HandlerOptions) Handler
}

// NewBaseListener creates a BaseListener from ListenerOptions
func NewBaseListener(options ListenerOptions) BaseListener {
	return BaseListener{
		closingMutex:   &sync.Mutex{},
		Config:         options.ListenerConfig,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}

// GetConfig implements plugin_v1.Listener
func (l *BaseListener) GetConfig() v1.Listener {
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

// GetName returns the internal name given to this listener
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
	// If we are already in the process of shutdown, don't try to
	// shutdown again. Wrap this check in a sync lock to prevent
	// race conditions
	l.closingMutex.Lock()

	if l.IsClosed {
		return nil
	}
	l.IsClosed = true

	l.closingMutex.Unlock()

	log.Printf("Shutting down listener's handlers...")

	var err error
	var wg sync.WaitGroup

	for _, handler := range l.handlers {
		// block scoped variable for use in goroutine
		_handler := handler

		wg.Add(1)
		go func() {
			defer wg.Done()
			_handler.Shutdown()
		}()
	}

	wg.Wait()

	if l.NetListener != nil {
		err = l.NetListener.Close()
		l.NetListener = nil
	}

	return err
}

// AddHandler appends a given Handler to the slice of Handlers held by BaseListener
func (l *BaseListener) AddHandler(handler Handler) {
	if l.handlers == nil {
		l.handlers = make([]Handler, 0)
	}

	l.handlers = append(l.handlers, handler)
}

// RemoveHandler removes a given Handler from the slice of Handlers held by BaseListener
func (l *BaseListener) RemoveHandler(targetHandler Handler) {
	var handlers []Handler
	for _, handler := range l.handlers {
		if handler == targetHandler {
			continue
		} else {
			handlers = append(handlers, handler)
		}
	}

	l.handlers = handlers
}
