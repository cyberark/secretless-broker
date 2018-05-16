package mysql

import (
	"log"
	"net"

	"github.com/conjurinc/secretless/internal/app/secretless/mysql/protocol"
	"github.com/conjurinc/secretless/pkg/secretless/config"
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
	Config        config.Handler
	Client        net.Conn
	Backend       net.Conn
	BackendConfig *BackendConfig
}

func (h *Handler) abort(err error) {
	mysqlError := protocol.Error{
		Code:     protocol.CRUnknownError,
		SQLSTATE: protocol.ErrorCodeInternalError,
		Message:  err.Error(),
	}
	h.Client.Write(mysqlError.GetMessage())
}

func stream(source, dest net.Conn) {
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
	}
}

// Pipe performs continuous bidirectional transfer of data between the client and backend.
func (h *Handler) Pipe() {
	if h.Config.Debug {
		log.Printf("Connecting client %s to backend %s", h.Client.RemoteAddr(), h.Backend.RemoteAddr())
	}

	go stream(h.Client, h.Backend)
	go stream(h.Backend, h.Client)
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
