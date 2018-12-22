package example

import (
	"log"
	"net"
	"strconv"
)

func (h *Handler) abort(err error) {
	h.GetClientConnection().Write([]byte("Error"))
}

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
// TODO: what is the file doing here?  what is this an example of?
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{}

	var values map[string][]byte
	if values, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		log.Fatalf("FATAL! Could not resolve all credentials for example plugin!\n%s", err)
		return
	}

	log.Println("Example plugin variables resolved")

	if h.GetConfig().Debug {
		log.Printf("Example backend connection parameters: %s", values)
	}

	if host := values["host"]; host != nil {
		result.Host = string(values["host"])
	}

	if values["port"] != nil {
		port64, _ := strconv.ParseUint(string(values["port"]), 10, 64)
		result.Port = uint(port64)
	}

	result.ProviderVariable = string(values["providerVariable"])

	delete(values, "host")
	delete(values, "port")
	delete(values, "providerVariable")

	h.BackendConfig = &result
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

	if h.GetConfig().Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	h.BackendConnection = backend

	return
}
