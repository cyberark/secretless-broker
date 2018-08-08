package pg

import (
	"log"
	"net"
	"io"
	"fmt"


	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
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
	plugin_v1.BaseHandler
	BackendConfig    *BackendConfig
	ClientOptions    *ClientOptions
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
		log.Printf("Connecting client %s to backend %s", h.GetClientConnection().RemoteAddr(), h.GetBackendConnection().RemoteAddr())
	}

	go stream(h.GetClientConnection(), h.GetBackendConnection(), func(b []byte) {
		h.EventNotifier.ClientData(h.ClientConnection, b)
	})
	go stream(h.GetBackendConnection(), h.GetClientConnection(), func(b []byte) {
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

func (h *Handler) Shutdown() {
	defer h.BaseHandler.Shutdown()

	h.abort(fmt.Errorf("secretless shutting down"))
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}

	handler.Run()

	return handler
}
