package example

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// BackendConfig stores the connection info to the real backend database.
// These values are pulled from the handler credentials config
type BackendConfig struct {
	Host string
	Port uint
}

// Handler connects a client to a backend. It uses the handler Config and Providers to
// establish the BackendConfig, which is used to make the Backend connection. Then the data
// is transferred bidirectionally between the Client and Backend.
type Handler struct {
	Backend          net.Conn
	BackendConfig    *BackendConfig
	ClientConnection net.Conn
	HandlerConfig    config.Handler
	EventNotifier    plugin_v1.EventNotifier
}

func stream(source, dest net.Conn, callback func([]byte)) {
	timeoutDuration := 2 * time.Second
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {
		source.SetReadDeadline(time.Now().Add(timeoutDuration))
		length, err = source.Read(buffer)
		if err != nil {
			dest.Close()
			source.Close()
			return
		}

		dest.SetReadDeadline(time.Now().Add(timeoutDuration))
		_, err = dest.Write(buffer[:length])
		if err != nil {
			dest.Close()
			source.Close()
			return
		}

		callback(buffer[:length])
	}
}

func streamWithTransform(source, dest net.Conn, callback func([]byte)) {
	timeoutDuration := 2 * time.Second
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {
		source.SetReadDeadline(time.Now().Add(timeoutDuration))
		length, err = source.Read(buffer)
		if err != nil {
			dest.Close()
			source.Close()
			return
		}

		lines := strings.Split(string(buffer), "\r\n")
		newLines := append(lines, "")

		insertIndex := 2
		copy(newLines[insertIndex+1:], newLines[insertIndex:])
		newLines[insertIndex] = "Example-Header: IsSet"

		newContent := strings.Join(newLines, "\r\n")

		dest.SetReadDeadline(time.Now().Add(timeoutDuration))
		_, err = dest.Write(bytes.NewBufferString(newContent).Bytes())
		if err != nil {
			dest.Close()
			source.Close()
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

	go streamWithTransform(h.GetClientConnection(), h.Backend, func(b []byte) {
		h.EventNotifier.ClientData(h.GetClientConnection(), b)
	})
	go stream(h.Backend, h.GetClientConnection(), func(b []byte) {
		h.EventNotifier.ClientData(h.GetClientConnection(), b)
	})
}

// Run configures the backend connection info, connects to the backend to
// complete the connection phase, and pipes the data between the client and
// the backend
func (h *Handler) Run() {
	var err error

	if err = h.ConfigureBackend(); err != nil {
		log.Println("Configuring backend failed!")
		h.abort(err)
		return
	}

	if err = h.ConnectToBackend(); err != nil {
		log.Println("Connecting to backend failed!")
		h.abort(err)
		return
	}

	if h.EventNotifier == nil {
		h.abort(errors.New("ERROR! EventNotifier was not set in example handler"))
		return
	}

	h.Pipe()
}

// GetConfig implements secretless.Handler
func (h *Handler) GetConfig() config.Handler {
	return h.HandlerConfig
}

// Authenticate is not used here
// TODO: Remove this when interface is cleaned up
func (h *Handler) Authenticate(map[string][]byte, *http.Request) error {
	return errors.New("example listener does not use Authenticate")
}

// GetClientConnection implements secretless.Handler
func (h *Handler) GetClientConnection() net.Conn {
	return h.ClientConnection
}

// GetBackendConnection implements secretless.Handler
func (h *Handler) GetBackendConnection() net.Conn {
	return nil
}

// LoadKeys is not used here
// TODO: Remove this when interface is cleaned up
func (h *Handler) LoadKeys(keyring agent.Agent) error {
	return errors.New("example handler does not use LoadKeys")
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
