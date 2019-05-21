package pg

import (
	"fmt"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/ssl"
	"log"
	"net"
	"net/url"
	"reflect"

	"github.com/cyberark/secretless-broker/internal/app/secretless/listeners/pg/protocol"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
)

var sslOptions = []string{
	"sslrootcert",
	"sslmode",
	"sslkey",
	"sslcert",
}

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// connectionDetails field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{
		Options:      make(map[string]string),
		QueryStrings: make(map[string]string),
	}

	var values map[string][]byte
	if values, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return
	}

	if h.GetConfig().Debug {
		keys := reflect.ValueOf(values).MapKeys()
		log.Printf("%s backend connection parameters: %s", h.GetConfig().Name, keys)
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

	for _, sslOption := range sslOptions {
		if values[sslOption] != nil {
			value := string(values[sslOption])
			if value != "" {
				result.QueryStrings[sslOption] = value
			}
		}
		delete(values, sslOption)
	}

	delete(values, "address")
	delete(values, "username")
	delete(values, "password")

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

	tlsConf, err := ssl.NewDbSSLMode(h.BackendConfig.QueryStrings, true)
	if err != nil {
		return
	}

	if tlsConf.UseTLS {
		// Start SSL Check
		/*
		 * First determine if SSL is allowed by the backend. To do this, send an
		 * SSL request. The response from the backend will be a single byte
		 * message. If the value is 'S', then SSL connections are allowed and an
		 * upgrade to the connection should be attempted. If the value is 'N',
		 * then the backend does not support SSL connections.
		 */

		/* Create the SSL request message. */
		message := protocol.NewMessageBuffer([]byte{})
		message.WriteInt32(8)
		message.WriteInt32(protocol.SSLRequestCode)

		/* Send the SSL request message. */
		_, err := connection.Write(message.Bytes())

		if err != nil {
			return err
		}

		/* Receive SSL response message. */
		response := make([]byte, 4096)
		_, err = connection.Read(response)

		if err != nil {
			return err
		}

		/*
		 * If SSL is not allowed by the backend then close the connection and
		 * throw an error.
		 */
		if len(response) > 0 && response[0] != 'S' {
			fmt.Println(string(response))
			connection.Close()
			return fmt.Errorf("the backend does not allow SSL connections")
		}
		// End SSL Check

		// Accept renegotiation requests initiated by the backend.
		//
		// Renegotiation was deprecated then removed from PostgreSQL 9.5, but
		// the default configuration of older versions has it enabled. Redshift
		// also initiates renegotiations and cannot be reconfigured.
		// Switch to TLS
		connection, err = ssl.HandleSSLUpgrade(connection, tlsConf)
		if err != nil {
			return err
		}
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
