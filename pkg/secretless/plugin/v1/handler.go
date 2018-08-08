package v1

import (
	"net"
	"net/http"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

//
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
// TODO: Remove LoadKeys as it's only used by sshagent listener
type Handler interface {
	Authenticate(map[string][]byte, *http.Request) error
	GetConfig() config.Handler
	GetClientConnection() net.Conn
	GetBackendConnection() net.Conn
	LoadKeys(keyring agent.Agent) error
	Shutdown()
}

type BaseHandler struct {
	self			   Handler
	HandlerConfig      config.Handler
	Resolver           Resolver
	EventNotifier      EventNotifier
	BackendConnection  net.Conn
	ClientConnection   net.Conn
	ShutdownNotifier   HandlerShutdownNotifier
}

func NewBaseHandler(options HandlerOptions) BaseHandler {
	return BaseHandler{
		HandlerConfig:     options.HandlerConfig,
		Resolver:          options.Resolver,
		EventNotifier:     options.EventNotifier,
		ClientConnection:  options.ClientConnection,
		ShutdownNotifier:  options.ShutdownNotifier,
	}
}

// Authenticate implements plugin_v1.Handler
func (h *BaseHandler) Authenticate(map[string][]byte, *http.Request) error {
	panic("BaseHandler does not implement Authenticate")
}

// GetConfig implements plugin_v1.Handler
func (h *BaseHandler) GetConfig() config.Handler {
	return h.HandlerConfig
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

// PipeHandlerWithStream performs continuous bidirectional transfer of data between handler client and backend
// stream: function for transfer
// eventNotifier: function ingesting transfer bytes
// handler: holder of client and backend connections
//
func PipeHandlerWithStream(handler Handler, stream func(net.Conn, net.Conn, func(b []byte)), eventNotifier EventNotifier, callback func()) {
	if handler.GetConfig().Debug {
		log.Printf("Connecting client %s to backend %s", handler.GetClientConnection().RemoteAddr(), handler.GetBackendConnection().RemoteAddr())
	}

	var _once sync.Once
	callbackOnce := func() {
		_once.Do(func() {
			callback()
		})
	}

	go func() {
		defer callbackOnce()
		stream(handler.GetClientConnection(), handler.GetBackendConnection(), func(b []byte) {
			eventNotifier.ClientData(handler.GetClientConnection(), b)
		})
	}()

	go func() {
		defer callbackOnce()
		stream(handler.GetBackendConnection(), handler.GetClientConnection(), func(b []byte) {
			eventNotifier.ServerData(handler.GetClientConnection(), b)
		})
	}()

}