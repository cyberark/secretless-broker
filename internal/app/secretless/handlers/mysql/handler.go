package mysql

import (
	"errors"
	"log"
	"net"
	"net/http"

	"golang.org/x/crypto/ssh/agent"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// BackendConfig stores the connection info to the real backend database.
// These values are pulled from the handler credentials config
type BackendConfig struct {
	Host     string
	Port     uint
	Username string
	Password string
	Options  map[string]string
}

// Handler connects a client to a backend. It uses the handler Config and Providers to
// establish the BackendConfig, which is used to make the Backend connection. Then the data
// is transferred bidirectionally between the Client and Backend.
//
// Handler requires "host", "port", "username" and "password" credentials.
type Handler struct {
	Backend          net.Conn
	BackendConfig    *BackendConfig
	ClientConnection net.Conn
	EventNotifier    plugin_v1.EventNotifier
	HandlerConfig    config.Handler
	Resolver         plugin_v1.Resolver
}

func (h *Handler) abort(err error) {
	mysqlError := protocol.Error{
		Code:     protocol.CRUnknownError,
		SQLSTATE: protocol.ErrorCodeInternalError,
		Message:  err.Error(),
	}
	h.GetClientConnection().Write(mysqlError.GetMessage())
}

func stream(source, dest net.Conn, callback func([]byte)) {
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {
		length, err = source.Read(buffer)
		if err != nil {
			return
		}

		_, err = dest.Write(buffer[:length])
		if err != nil {
			return
		}

		callback(buffer[:length])
	}
}

// Pipe performs continuous bidirectional transfer of data between the client and backend.
func (h *Handler) Pipe() {
	if h.GetConfig().Debug {
		log.Printf("Connecting client %s to backend %s", h.GetClientConnection().RemoteAddr(), h.Backend.RemoteAddr())
	}

	go stream(h.GetClientConnection(), h.Backend, func(b []byte) { h.EventNotifier.ClientData(h.GetClientConnection(), b) })
	go stream(h.Backend, h.GetClientConnection(), func(b []byte) { h.EventNotifier.ClientData(h.GetClientConnection(), b) })
}

// Run configures the backend connection info, connects to the backend to
// complete the connection phase, and pipes the data between the client and
// the backend
func (h *Handler) Run() {
	var err error

	if err = h.ConfigureBackend(); err != nil {
		h.abort(err)
		return
	}

	if err = h.ConnectToBackend(); err != nil {
		h.abort(err)
		return
	}

	h.Pipe()
}

// Authenticate is unused here
// TODO: Remove this when interface is cleaned up
func (h *Handler) Authenticate(map[string][]byte, *http.Request) error {
	return errors.New("mysql listener does not use Authenticate")
}

// GetConfig implements secretless.Handler
func (h *Handler) GetConfig() config.Handler {
	return h.HandlerConfig
}

// GetClientConnection implements secretless.Handler
func (h *Handler) GetClientConnection() net.Conn {
	return h.ClientConnection
}

// GetBackendConnection implements secretless.Handler
func (h *Handler) GetBackendConnection() net.Conn {
	return h.Backend
}

// LoadKeys is unused here
// TODO: Remove this when interface is cleaned up
func (h *Handler) LoadKeys(keyring agent.Agent) error {
	return errors.New("mysql handler does not use LoadKeys")
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		ClientConnection: options.ClientConnection,
		EventNotifier:    options.EventNotifier,
		HandlerConfig:    options.HandlerConfig,
		Resolver:         options.Resolver,
	}

	handler.Run()

	return handler
}
