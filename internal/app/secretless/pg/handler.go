package pg

import (
	"log"
	"net"

	"github.com/conjurinc/secretless/internal/app/secretless/pg/protocol"
	"github.com/conjurinc/secretless/internal/pkg/plugin"
	"github.com/conjurinc/secretless/pkg/secretless/config"
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
	Config        config.Handler
	Client        net.Conn
	Backend       net.Conn
	ClientOptions *ClientOptions
	BackendConfig *BackendConfig
}

func (h *Handler) abort(err error) {
	pgError := protocol.Error{
		Severity: protocol.ErrorSeverityFatal,
		Code:     protocol.ErrorCodeInternalError,
		Message:  err.Error(),
	}
	h.Client.Write(pgError.GetMessage())
}

func stream(source, dest net.Conn, callback func(net.Conn, []byte)) {
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

		callback(source, buffer[:length])
	}
}

// Pipe performs continuous bidirectional transfer of data between the client and backend.
func (h *Handler) Pipe() {
	if h.Config.Debug {
		log.Printf("Connecting client %s to backend %s", h.Client.RemoteAddr(), h.Backend.RemoteAddr())
	}

	go stream(h.Client, h.Backend, plugin.GetManager().ClientData)
	go stream(h.Backend, h.Client, plugin.GetManager().ServerData)
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

// GetConfig implements secretless.Handler
func (h *Handler) GetConfig() config.Handler {
	return h.Config
}

// GetClientConnection implements secretless.Handler
func (h *Handler) GetClientConnection() net.Conn {
	return h.Client
}

// GetBackendConnection implements secretless.Handler
func (h *Handler) GetBackendConnection() net.Conn {
	return h.Backend
}
