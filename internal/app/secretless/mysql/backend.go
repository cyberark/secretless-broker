package mysql

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/conjurinc/secretless/internal/app/secretless/mysql/protocol"
	"github.com/conjurinc/secretless/internal/app/secretless/variable"
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}

	var values map[string][]byte
	if values, err = variable.Resolve(h.Config.Credentials); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("MySQL backend connection parameters: %s", values)
	}

	if host := values["host"]; host != nil {
		// Form of url is : 'dbcluster.myorg.com:5432/reports'
		tokens := strings.SplitN(string(host), "/", 2)
		result.Host = tokens[0]
		if len(tokens) == 2 {
			result.Database = tokens[1]
		}
	}

	if values["port"] != nil {
		port64, _ := strconv.ParseUint(string(values["port"]), 10, 64)
		result.Port = uint(port64)
	}

	if values["username"] != nil {
		result.Username = string(values["username"])
	}

	if values["password"] != nil {
		result.Password = string(values["password"])
	}

	delete(values, "host")
	delete(values, "port")
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

	address := h.BackendConfig.Host + ":" + strconv.FormatUint(uint64(h.BackendConfig.Port), 10)

	if connection, err = net.Dial("tcp", address); err != nil {
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
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	if _, err = h.Client.Write(protocol.CreateAuthenticationOKMessage()); err != nil {
		return
	}

	h.Backend = connection

	return
}
