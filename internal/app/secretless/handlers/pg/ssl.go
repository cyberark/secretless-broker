package pg

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
	"net"
)

// Common error types
var (
	ErrSSLNotSupported           = errors.New("pq: SSL is not enabled on the server")
	ErrSSLKeyHasWorldPermissions = errors.New("pq: Private key file has group or world access. Permissions should be u=rw (0600) or less")
)
type options map[string]string

// ssl generates a function to upgrade a net.Conn based on the "sslmode" and
// related settings. The function is nil when no upgrade should take place.
func ssl(connection net.Conn, o options) (net.Conn, error) {
	tlsConf := tls.Config{
		InsecureSkipVerify: true,
	}

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
		return nil, err
	}

	/* Receive SSL response message. */
	response := make([]byte, 4096)
	_, err = connection.Read(response)

	if err != nil {
		return nil, err
	}

	/*
	 * If SSL is not allowed by the backend then close the connection and
	 * throw an error.
	 */
	if len(response) > 0 && response[0] != 'S' {
		fmt.Println(string(response))
		connection.Close()
		return nil, fmt.Errorf("the backend does not allow SSL connections")
	}
	// End SSL Check

	// Accept renegotiation requests initiated by the backend.
	//
	// Renegotiation was deprecated then removed from PostgreSQL 9.5, but
	// the default configuration of older versions has it enabled. Redshift
	// also initiates renegotiations and cannot be reconfigured.
	tlsConf.Renegotiation = tls.RenegotiateFreelyAsClient

	client := tls.Client(connection, &tlsConf)
	return client, nil
}
