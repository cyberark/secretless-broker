package mysql

import (
	"log"
	"net"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
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
	plugin_v1.BaseHandler
	BackendConfig    *BackendConfig
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
		log.Printf("Connecting client %s to backend %s", h.GetClientConnection().RemoteAddr(), h.GetBackendConnection().RemoteAddr())
	}

	go stream(h.GetClientConnection(), h.GetBackendConnection(), func(b []byte) { h.EventNotifier.ClientData(h.GetClientConnection(), b) })
	go stream(h.GetBackendConnection(), h.GetClientConnection(), func(b []byte) { h.EventNotifier.ClientData(h.GetClientConnection(), b) })
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
