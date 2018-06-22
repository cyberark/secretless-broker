package pg

import (
	"log"
	"net"
	"strings"

	"github.com/conjurinc/secretless/internal/app/secretless/listeners/pg/protocol"
	"github.com/conjurinc/secretless/internal/app/secretless/variable"
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}

	var values map[string][]byte
	if values, err = variable.Resolve(h.Config.Credentials, h.EventNotifier); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("PG backend connection parameters: %s", values)
	}

	if address := values["address"]; address != nil {
		// Form of url is : 'dbcluster.myorg.com:5432/reports'
		tokens := strings.SplitN(string(address), "/", 2)
		result.Address = tokens[0]
		if len(tokens) == 2 {
			result.Database = tokens[1]
		}
	}

	if values["username"] != nil {
		result.Username = string(values["username"])
	}
	if values["password"] != nil {
		result.Password = string(values["password"])
	}

	delete(values, "address")
	delete(values, "username")
	delete(values, "password")

	result.Options = make(map[string]string)
	for k, v := range values {
		result.Options[k] = string(v)
	}

	h.BackendConfig = &result

	return
}

// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {
	var connection net.Conn

	if connection, err = net.Dial("tcp", h.BackendConfig.Address); err != nil {
		return
	}

	if h.Config.Debug {
		log.Print("Sending startup message")
	}
	startupMessage := protocol.CreateStartupMessage(h.BackendConfig.Username, h.ClientOptions.Database, h.BackendConfig.Options)

	connection.Write(startupMessage)

	if h.Config.Debug {
		log.Print("Authenticating to the backend")
	}
	if err = protocol.HandleAuthenticationRequest(h.BackendConfig.Username, h.BackendConfig.Password, connection); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("Successfully connected to '%s'", h.BackendConfig.Address)
	}

	if _, err = h.Client.Write(protocol.CreateAuthenticationOKMessage()); err != nil {
		return
	}

	h.Backend = connection

	return
}
