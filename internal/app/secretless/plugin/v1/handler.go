package v1

import (
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

// HandlerShutdownNotifier is a function signature for notifying of a Handler's Shutdown
type HandlerShutdownNotifier func(Handler)

// HandlerOptions contains the configuration for the handler
type HandlerOptions struct {
	HandlerConfig    config_v1.Handler
	Channels         <-chan ssh.NewChannel
	ClientConnection net.Conn
	EventNotifier    EventNotifier
	ShutdownNotifier HandlerShutdownNotifier
	Resolver         Resolver
}

// Handler is an interface which takes a connection and connects it to a backend
// TODO: Remove Authenticate as it's only used by http listener
// TODO: Remove LoadKeys as it's only used by sshagent listener
type Handler interface {
	Authenticate(map[string][]byte, *http.Request) error
	GetConfig() config_v1.Handler
	GetClientConnection() net.Conn
	GetBackendConnection() net.Conn
	LoadKeys(keyring agent.Agent) error
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
	BackendConnection net.Conn
	ClientConnection  net.Conn
	EventNotifier     EventNotifier
	HandlerConfig     config_v1.Handler
	Resolver          Resolver
	ShutdownNotifier  HandlerShutdownNotifier
}

// NewBaseHandler creates a BaseHandler from HandlerOptions
func NewBaseHandler(options HandlerOptions) BaseHandler {
	return BaseHandler{
		ClientConnection: options.ClientConnection,
		EventNotifier:    options.EventNotifier,
		HandlerConfig:    options.HandlerConfig,
		Resolver:         options.Resolver,
		ShutdownNotifier: options.ShutdownNotifier,
	}
}

// Authenticate implements plugin_v1.Handler
func (h *BaseHandler) Authenticate(map[string][]byte, *http.Request) error {
	panic("BaseHandler does not implement Authenticate")
}

// GetConfig implements plugin_v1.Handler
func (h *BaseHandler) GetConfig() config_v1.Handler {
	return h.HandlerConfig
}

// Debug prints the given msg, but only if Debug is enabled.
func (h *BaseHandler) Debug(msg string) {
	if h.DebugModeOn() {
		log.Print(msg)
	}
}

// Debugf prints the given msg, but only if Debug is enabled.
func (h *BaseHandler) Debugf(format string, v ...interface{}) {
	if h.DebugModeOn() {
		log.Printf(format, v...)
	}
}

// DebugModeOn tells you if debug mode is enabled.
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

// LoadKeys implements plugin_v1.Handler
func (h *BaseHandler) LoadKeys(keyring agent.Agent) error {
	panic("BaseHandler does not implement LoadKeys")
}

// Shutdown implements plugin_v1.Handler
func (h *BaseHandler) Shutdown() {
	log.Printf("Handler shutting down...")
	h.ShutdownNotifier(h)
}

// PipeHandlerWithStream performs continuous bidirectional transfer of data
// between handler client and backend takes arguments:
//
// [handler]:
//   Handler-compliant struct. Handler#GetClientConnection and
//   Handler#GetBackendConnection provide client and backend connections
//   (net.Conn)
// [stream]:
//   function performing continuous bidirectional transfer
// [eventNotifier]:
//   EventNotifier-compliant struct. EventNotifier#ClientData is passed transfer
//   bytes
// [done]:
//   function called once when transfer ceases
//
// NOTE:
//
// This function is a confusing and probably the wrong abstraction.  This
// function exists only to trigger its done() callback, and seems to have been
// created only DRY up some code repeated across pg and mysql handlers.
//
// It accepts a general "stream" function.  That does the actual work of
// streaming.  And stream accepts a callback for the sole purpose of triggering
// events on the eventNotifier.
//
// Some other funny stuff:
//   - Even though eventNotifier is a property of the handlers we use this for
//     (pg, mysql) we have to pass it in separately because it's not a method
//     on the Handler _interface_.
//   - Ditto the above for the passed in "done" func, which in both cases is
//     just the ShutdownNotifier, also available from the handler.
//   - This is an unnatural partitioning, and it's _screaming_ at us that we've
//     done something wrong.
//   - "stream" is _exactly_ the same in both handlers
//
//
func PipeHandlerWithStream(
	handler Handler,
	stream func(net.Conn, net.Conn, func(b []byte)),
	eventNotifier EventNotifier,
	done func(),
) {
	if handler.GetConfig().Debug {
		log.Printf(
			"Connecting client %s to backend %s",
			handler.GetClientConnection().RemoteAddr(),
			handler.GetBackendConnection().RemoteAddr(),
		)
	}

	// Q: Why isn't this just:
	//     _once.Do(done)
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
