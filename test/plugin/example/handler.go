package example

import (
	"bytes"
	"errors"
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/agent"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// connectionDetails stores the connection info to the real backend database.
// These values are pulled from the handler credentials config
type BackendConfig struct {
	Host             string
	Port             uint
	ProviderVariable string
}

// Handler connects a client to a backend. It uses the handler Config and Providers to
// establish the connectionDetails, which is used to make the Backend connection. Then the data
// is transferred bidirectionally between the Client and Backend.
type Handler struct {
	BackendConfig     *BackendConfig
	BackendConnection net.Conn
	ClientConnection  net.Conn
	EventNotifier     plugin_v1.EventNotifier
	HandlerConfig     v1.Handler
	Resolver          plugin_v1.Resolver
	ShutdownNotifier  plugin_v1.HandlerShutdownNotifier
}

func stream(source, dest net.Conn, callback func([]byte)) {
	defer func() {
		source.Close()
		dest.Close()
	}()

	timeoutDuration := 2 * time.Second
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {
		source.SetReadDeadline(time.Now().Add(timeoutDuration))
		length, err = source.Read(buffer)
		if err != nil {
			return
		}

		dest.SetReadDeadline(time.Now().Add(timeoutDuration))
		_, err = dest.Write(buffer[:length])
		if err != nil {
			return
		}

		callback(buffer[:length])
	}
}

func streamWithTransform(source, dest net.Conn, config *BackendConfig, callback func([]byte)) {
	defer func() {
		source.Close()
		dest.Close()
	}()

	timeoutDuration := 2 * time.Second
	buffer := make([]byte, 4096)

	var length int
	var err error

	for {
		source.SetReadDeadline(time.Now().Add(timeoutDuration))
		length, err = source.Read(buffer)
		if err != nil {
			return
		}

		lines := strings.Split(string(buffer), "\r\n")
		newLines := append(lines, "", "")

		insertIndex := 2
		copy(newLines[insertIndex+2:], newLines[insertIndex:])
		newLines[insertIndex] = "Example-Header: IsSet"
		newLines[insertIndex+1] = "Example-Provider-Secret: " + config.ProviderVariable

		newContent := strings.Join(newLines, "\r\n")

		dest.SetReadDeadline(time.Now().Add(timeoutDuration))
		_, err = dest.Write(bytes.NewBufferString(newContent).Bytes())
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

	go streamWithTransform(h.GetClientConnection(), h.GetBackendConnection(), h.BackendConfig, func(b []byte) {
		h.EventNotifier.ClientData(h.GetClientConnection(), b)
	})
	go stream(h.GetBackendConnection(), h.GetClientConnection(), func(b []byte) {
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

// Authenticate implements plugin_v1.Handler
func (h *Handler) Authenticate(map[string][]byte, *http.Request) error {
	panic("example handler does not implement Authenticate")
}

// GetConfig implements plugin_v1.Handler
func (h *Handler) GetConfig() v1.Handler {
	return h.HandlerConfig
}

// GetClientConnection implements plugin_v1.Handler
func (h *Handler) GetClientConnection() net.Conn {
	return h.ClientConnection
}

// GetBackendConnection implements plugin_v1.Handler
func (h *Handler) GetBackendConnection() net.Conn {
	return h.BackendConnection
}

// LoadKeys implements plugin_v1.Handler
func (h *Handler) LoadKeys(keyring agent.Agent) error {
	panic("example handler does not implement LoadKeys")
}

// Shutdown implements plugin_v1.Handler
func (h *Handler) Shutdown() {
	log.Printf("example handler shutting down...")
	h.ShutdownNotifier(h)
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		ClientConnection:  options.ClientConnection,
		EventNotifier:     options.EventNotifier,
		HandlerConfig:     options.HandlerConfig,
		Resolver:          options.Resolver,
		ShutdownNotifier:  options.ShutdownNotifier,
	}

	handler.Run()

	return handler
}
