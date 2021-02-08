package pg

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/ssl"
)

// handleSSL conditionally upgrades the backend connection to SSL depending on the
// connection details.
func (s *SingleUseConnector) handleSSL() error {
	tlsConf, err := ssl.NewDbSSLMode(
		s.connectionDetails.SSLOptions,
		true)
	if err != nil {
		return err
	}

	if !tlsConf.UseTLS {
		return nil
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
	_, err = s.backendConn.Write(message.Bytes())
	if err != nil {
		return err
	}

	/* Receive SSL response message. */
	response := make([]byte, 4096)
	_, err = s.backendConn.Read(response)

	if err != nil {
		return err
	}

	/*
	 * If SSL is not allowed by the backend then close the connection and
	 * throw an error.
	 */
	if len(response) > 0 && response[0] != 'S' {
		s.backendConn.Close()
		return fmt.Errorf("the backend does not allow SSL connections")
	}
	// End SSL Check

	// Accept renegotiation requests initiated by the backend.
	//
	// Renegotiation was deprecated then removed from PostgreSQL 9.5, but
	// the default configuration of older versions has it enabled. Redshift
	// also initiates renegotiations and cannot be reconfigured.
	// Switch to TLS
	s.backendConn, err = ssl.HandleSSLUpgrade(s.backendConn, tlsConf)
	return err
}

// ConnectToBackend establishes the connection to the target database and sets
// the backendConnection field.
func (s *SingleUseConnector) ConnectToBackend() error {
	var err error

	s.backendConn, err = net.Dial("tcp", s.connectionDetails.Address())
	if err != nil {
		return err
	}

	s.logger.Debugln("Sending startup message")

	err = s.handleSSL()
	if err != nil {
		return err
	}

	startupMessage := protocol.CreateStartupMessage(
		protocol.ProtocolVersion,
		s.connectionDetails.Username,
		s.databaseName,
		s.connectionDetails.Options,
	)

	s.backendConn.Write(startupMessage)

	s.logger.Debugln("Authenticating to the backend")

	err = protocol.HandleAuthenticationRequest(
		s.connectionDetails.Username,
		s.connectionDetails.Password,
		s.backendConn)
	if err != nil {
		return err
	}

	s.logger.Debugf(
		"Successfully connected to '%s:%s'",
		s.connectionDetails.Host,
		s.connectionDetails.Port,
	)

	if _, err = s.clientConn.Write(protocol.CreateAuthenticationOKMessage()); err != nil {
		return err
	}

	return nil
}
