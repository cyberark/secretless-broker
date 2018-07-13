package pg

import (
	"errors"
	"log"
	"net"
	"net/http"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/internal/app/secretless/handlers/pg/protocol"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	"io"
)

// ClientOptions stores the option that were specified by the connection client.
// The User and Database are required. Other options are stored in a map.
type ClientOptions struct {
	// TODO: remove this when custom authorization is removed.
	User string
	// TODO: override the database address with this setting.
	Database string
	Options  map[string]string
}

// BackendConfig stores the connection info to the real backend database.
type BackendConfig struct {
	Address  string
	Username string
	Password string
	Database string
	Options  map[string]string
}

// Handler connects a client to a backend. It uses the handler Config and Providers to
// establish the BackendConfig, which is used to make the Backend connection. Then the data
// is transferred bidirectionally between the Client and Backend.
//
// Handler requires "address", "username" and "password" credentials.
type Handler struct {
	Backend          net.Conn
	BackendConfig    *BackendConfig
	HandlerConfig    config.Handler
	ClientConnection net.Conn
	ClientOptions    *ClientOptions
	EventNotifier    plugin_v1.EventNotifier
}

func (h *Handler) abort(err error) {
	pgError := protocol.Error{
		Severity: protocol.ErrorSeverityFatal,
		Code:     protocol.ErrorCodeInternalError,
		Message:  err.Error(),
	}
	h.GetClientConnection().Write(pgError.GetMessage())
}

func stream(source, dest net.Conn, callback func([]byte)) {
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {

		length, err = source.Read(buffer)
		if err != nil {
			if err == io.EOF {
				source.Close()
				dest.Close()

				log.Printf("source %s closed for destination %s", source.RemoteAddr(), dest.RemoteAddr())
			}
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

	go stream(h.GetClientConnection(), h.Backend, func(b []byte) {
		h.EventNotifier.ClientData(h.ClientConnection, b)
	})
	go stream(h.Backend, h.GetClientConnection(), func(b []byte) {
		h.EventNotifier.ServerData(h.GetClientConnection(), b)
	})
}

// Run processes the startup message, configures the backend connection, connects to the backend,
// and pipes the data between the client and the backend.
func (h *Handler) Run() {
	var err error

	if err = h.Startup(); err != nil {
		h.abort(err)
		return
	}

	if err := h.ConfigureBackend(); err != nil {
		h.abort(err)
		return
	}

	if err = h.ConnectToBackend(); err != nil {
		h.abort(err)
		return
	}

	h.Pipe()
}

// Authenticate is not used here
// TODO: Remove this when interface is cleaned up
func (h *Handler) Authenticate(map[string][]byte, *http.Request) error {
	return errors.New("pg listener does not use Authenticate")
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

// LoadKeys is not used here
// TODO: Remove this when interface is cleaned up
func (h *Handler) LoadKeys(keyring agent.Agent) error {
	return errors.New("pg handler does not use LoadKeys")
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		ClientConnection: options.ClientConnection,
		EventNotifier:    options.EventNotifier,
		HandlerConfig:    options.HandlerConfig,
	}

	handler.Run()

	return handler
}
