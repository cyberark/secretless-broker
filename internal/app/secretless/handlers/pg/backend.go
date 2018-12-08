package pg

import (
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}
	result.Options = make(map[string]string)
	result.QueryStrings = make(map[string]string)

	var values map[string][]byte
	if values, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Printf("PG backend connection parameters: %s", values)
	}

	if address := values["address"]; address != nil {
		u, err := url.Parse(fmt.Sprintf("postgres://%s", address))
		if err != nil {
			return err
		}

		result.Address = u.Host
		result.Database = u.Path
		for k, v := range u.Query() {
			if len(v) > 0 {
				result.QueryStrings[k] = string(v[0])
			}
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

	// TODO: remove because of hack

	// TODO: what are options for, it's weird that any additonal
	// credentials are passed to postgres as options
	delete(values, "port")
	delete(values, "host")
	delete(values, "sslrootcert")
	delete(values, "sslmode")

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

	debug := util.OptionalDebug(h.GetConfig().Debug)
	debug("Sending startup message")

	connection, err = ssl(connection, h.BackendConfig.QueryStrings)
	if err != nil {
		return
	}

	startupMessage := protocol.CreateStartupMessage(h.BackendConfig.Username, h.ClientOptions.Database, h.BackendConfig.Options)

	connection.Write(startupMessage)

	debug("Authenticating to the backend")

	if err = protocol.HandleAuthenticationRequest(h.BackendConfig.Username, h.BackendConfig.Password, connection); err != nil {
		return
	}

	debug("Successfully connected to '%s'", h.BackendConfig.Address)

	if _, err = h.GetClientConnection().Write(protocol.CreateAuthenticationOKMessage()); err != nil {
		return
	}

	h.BackendConnection = connection

	return
}
