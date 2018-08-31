package passthrough

import (
	"io"
	"log"
	"net"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// BackendConfig stores the connection info to the real backend
type BackendConfig struct {
	Address string
}

// Handler stores the object for pointing to the appropriate backend server
type Handler struct {
	plugin_v1.BaseHandler
}

// Run performs continuous bidirectional transfer of data between the client and backend.
func (handler *Handler) Run() {
	if handler.GetConfig().Debug {
		log.Println("Resolving variables...")
	}

	backendVariables, err := handler.Resolver.Resolve(handler.GetConfig().Credentials)
	if err != nil {
		log.Fatalf("FATAL: Could not resolve passthrough handler varaibles!")
	}

	var backendAddress = string(backendVariables["address"])

	if handler.GetConfig().Debug {
		log.Println("Resolving variables done")
		log.Printf("Connecting to %s...", backendAddress)
	}

	conn, err := net.Dial("tcp", backendAddress)
	if err != nil {
		log.Fatalf("ERROR: Could not connect to backend '%s'!", backendAddress)
	}
	handler.BackendConnection = conn

	if handler.GetConfig().Debug {
		log.Printf("Connecting client %s to backend %s",
			handler.GetClientConnection().RemoteAddr(),
			handler.GetBackendConnection().RemoteAddr())
	}

	go func() {
		if _, err := io.Copy(handler.GetClientConnection(), handler.GetBackendConnection()); err != nil {
			log.Println("Closing from client")
			handler.Shutdown()
		}
	}()

	go func() {
		if _, err := io.Copy(handler.GetBackendConnection(), handler.GetClientConnection()); err != nil {
			log.Println("Closing from server")
			handler.Shutdown()
		}
	}()

}

// Shutdown tries to nicely close our connection
func (handler *Handler) Shutdown() {
	log.Println("Shutting down passthrough handler")
	defer handler.BaseHandler.Shutdown()

	handler.GetBackendConnection().Close()
	handler.GetClientConnection().Close()

	handler.ShutdownNotifier(handler)
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}

	handler.Run()

	return handler
}
