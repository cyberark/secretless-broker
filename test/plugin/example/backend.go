package example

import (
	"log"
	"net"
	"strconv"

	"github.com/conjurinc/secretless/internal/app/secretless/variable"
)

func (h *Handler) abort(err error) {
	h.Client.Write([]byte("Error"))
}

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{}

	var values map[string][]byte
	if values, err = variable.Resolve(h.Config.Credentials, h.EventNotifier); err != nil {
		log.Fatalf("FATAL! Could not resolve credentials!")
		return
	}

	if h.Config.Debug {
		log.Printf("Example backend connection parameters: %s", values)
	}

	if host := values["host"]; host != nil {
		result.Host = string(values["host"])
	}

	if values["port"] != nil {
		port64, _ := strconv.ParseUint(string(values["port"]), 10, 64)
		result.Port = uint(port64)
	}

	delete(values, "host")
	delete(values, "port")

	h.BackendConfig = &result
	log.Println(h.BackendConfig)

	return
}

// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {
	var backend net.Conn

	address := h.BackendConfig.Host + ":" + strconv.FormatUint(uint64(h.BackendConfig.Port), 10)
	log.Printf("Using backend '%s' for plugin test handler", address)

	if backend, err = net.Dial("tcp", address); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	h.Backend = backend

	return
}
