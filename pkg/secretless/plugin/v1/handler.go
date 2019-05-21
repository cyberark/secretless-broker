package v1

import (
	"log"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// HandlerShutdownNotifier is a function signature for notifying of a Handler's Shutdown
type HandlerShutdownNotifier func(Handler)

// HandlerOptions contains the configuration for the handler
type HandlerOptions struct {
	HandlerConfig           config.Handler
	Channels                <-chan ssh.NewChannel
	ClientConnection        net.Conn
	EventNotifier           EventNotifier
	ShutdownNotifier        HandlerShutdownNotifier
	Resolver                Resolver
}

// Handler is an interface which takes a connection and connects it to a backend
// TODO: Remove Authenticate as it's only used by http listener
type Handler interface {
	GetConfig() config.Handler
	GetClientConnection() net.Conn
	GetBackendConnection() net.Conn
	Shutdown()
}

// BaseHandler provides default (shared/common) implementations
// of Handler interface methods, where it makes sense
// - the rest of the methods panic if
// not implemented in the "DerivedHandler"
// e.g. BaseHandler#Authenticate.
//
// The intention is to keep things DRY by
// embedding BaseHandler in "DerivedHandler"
//
// There is no requirement to use BaseHandler.
type BaseHandler struct {
	BackendConnection  net.Conn
	ClientConnection   net.Conn
	EventNotifier      EventNotifier
	HandlerConfig      config.Handler
	Resolver           Resolver
	ShutdownNotifier   HandlerShutdownNotifier
}

// NewBaseHandler creates a BaseHandler from HandlerOptions
func NewBaseHandler(options HandlerOptions) BaseHandler {
	return BaseHandler{
		ClientConnection:  options.ClientConnection,
		EventNotifier:     options.EventNotifier,
		HandlerConfig:     options.HandlerConfig,
		Resolver:          options.Resolver,
		ShutdownNotifier:  options.ShutdownNotifier,
	}
}

// GetConfig implements plugin_v1.Handler
func (h *BaseHandler) GetConfig() config.Handler {
	return h.HandlerConfig
}

// Debug: Print only if Debug is enabled
func (h *BaseHandler) Debug(msg string) {
	if h.DebugModeOn() {
		log.Print(msg)
	}
}

// Debug: Print only if Debug is enabled
func (h *BaseHandler) Debugf(format string, v ...interface{}) {
	if h.DebugModeOn() {
		log.Printf(format, v...)
	}
}

func (h *BaseHandler) DebugModeOn() bool {
	return h.GetConfig().Debug
}

// GetClientConnection implements plugin_v1.Handler
func (h *BaseHandler) GetClientConnection() net.Conn {
	return h.ClientConnection
}

// GetBackendConnection implements plugin_v1.Handler
func (h *BaseHandler) GetBackendConnection() net.Conn {
	return h.BackendConnection
}

// Shutdown implements plugin_v1.Handler
func (h *BaseHandler) Shutdown() {
	log.Printf("Handler shutting down...")
	h.ShutdownNotifier(h)
}

// PipeHandlerWithStream performs continuous bidirectional transfer of data between handler client and backend
// takes arguments ->
// [handler]: Handler-compliant struct. Handler#GetClientConnection and Handler#GetBackendConnection provide client and backend connections (net.Conn)
// [stream]: function performing continuous bidirectional transfer
// [eventNotifier]: EventNotifier-compliant struct. EventNotifier#ClientData is passed transfer bytes
// [done]: function called once when transfer ceases
func PipeHandlerWithStream(handler Handler, stream func(net.Conn, net.Conn, func(b []byte)), eventNotifier EventNotifier, done func()) {
	if handler.GetConfig().Debug {
		log.Printf("Connecting client %s to backend %s", handler.GetClientConnection().RemoteAddr(), handler.GetBackendConnection().RemoteAddr())
	}

	var _once sync.Once
	doneOnce := func() {
		_once.Do(func() {
			done()
		})
	}

	go func() {
		defer doneOnce()
		stream(handler.GetClientConnection(), handler.GetBackendConnection(), func(b []byte) {
			eventNotifier.ClientData(handler.GetClientConnection(), b)
		})
	}()

	go func() {
		defer doneOnce()
		stream(handler.GetBackendConnection(), handler.GetClientConnection(), func(b []byte) {
			eventNotifier.ServerData(handler.GetClientConnection(), b)
		})
	}()

}