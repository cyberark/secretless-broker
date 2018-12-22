package pg

import (
	"fmt"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/ssl"
	"net"
	"net/url"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
)

//TODO move this to another file

type PgAddress struct {
	*url.URL
	SslMode string
	SslRootCert string
}

func NewPgAddress(address string) (*PgAddress, error) {
	url, err := url.Parse(fmt.Sprintf("postgres://%s", address))

	if err != nil {
		return nil, err
	}

	result := &PgAddress{URL: url, SslMode: "", SslRootCert: ""}

	for k, v := range url.Query() {
		if k == "sslmode" {
			result.SslMode = string(v[0])
		}
		if k == "sslrootcert" {
			result.SslRootCert = string(v[0])
		}
	}

	return result, nil
}


//func (url *string) (url.URL, error) {
//	if url == nil {
//		return errors.
//
//	}
//		u, err := url.Parse(fmt.Sprintf("postgres://%s", address))
//		if err != nil {
//			return err
//		}
//
//}

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() error {

	result := BackendConfig{}
	result.Options = make(map[string]string)
	result.QueryStrings = make(map[string]string)

	// Fetch credentials
	rawCredentials, err := h.Credentials()
	if err != nil {
		return err
	}

	// Convert to strings
	credentials := make(map[string]string)
	for k, v := range rawCredentials {
		credentials[k] = string(v)
	}

	h.Debugf("PG backend connection parameters: %s", credentials)

	// sslmode and sslrootcert are first taken from credentials
	// the overridden by the query string
	if address := rawCredentials["address"]; address != nil {
		u, err := url.Parse(fmt.Sprintf("postgres://%s", address))
		if err != nil {
			return err
		}
		pgAddress, err := NewPgAddress(string(address))

		result.Address = pgAddress.Host
		result.Database = pgAddress.Path
		for k, v := range u.Query() {
			if len(v) > 0 {
				result.QueryStrings[k] = string(v[0])
			}
		}
	}

	//TODO: Why were these previously surrounded by nil checks?
	//
	result.Username = credentials["username"]
	result.Password = credentials["password"]

	if rawCredentials["sslrootcert"] != nil {
		sslrootcert := credentials["sslrootcert"]
		if sslrootcert != "" {
			result.QueryStrings["sslrootcert"] = sslrootcert
		}
	}
	if rawCredentials["sslmode"] != nil {
		sslmode := credentials["sslmode"]
		if sslmode != "" {
			result.QueryStrings["sslmode"] = sslmode
		}
	}

	// Remove the keys we've already captured
	delete(credentials, "address")
	delete(credentials, "username")
	delete(credentials, "password")
	delete(credentials, "sslrootcert")
	delete(credentials, "sslmode")

	// Everything else is an "option"
	for k, v := range credentials {
		result.Options[k] = v
	}

	h.BackendConfig = &result

	return nil
}

// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {
	var connection net.Conn

	if connection, err = net.Dial("tcp", h.BackendConfig.Address); err != nil {
		return
	}

	h.Debugf("Sending startup message")

	tlsConf, err := ssl.NewSecretlessTLSConfig(h.BackendConfig.QueryStrings, true)
	if err != nil {
		return
	}

	//TODO: this does't belong here
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

	h.Debugf("Authenticating to the backend")

	if err = protocol.HandleAuthenticationRequest(h.BackendConfig.Username, h.BackendConfig.Password, connection); err != nil {
		return
	}

	h.Debugf("Successfully connected to '%s'", h.BackendConfig.Address)

	if _, err = h.GetClientConnection().Write(protocol.CreateAuthenticationOKMessage()); err != nil {
		return
	}

	h.BackendConnection = connection

	return
}
