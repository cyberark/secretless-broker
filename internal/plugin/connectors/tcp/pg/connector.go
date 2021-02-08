package pg

import (
	"io"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// SingleUseConnector is used to create an authenticated connection to a PostgreSQL target
// service using a client connection and connection details.
type SingleUseConnector struct {
	// The connections are decorated versions of net.Conn that allow us
	// to do read/writes according to the PostgreSQL protocol.  Protocol level
	// details are thus encapsulated. Within the PostgreSQL code, we _only_
	// deal with these decorated versions.

	clientConn        net.Conn
	backendConn       net.Conn
	logger            log.Logger
	connectionDetails *ConnectionDetails
	// databaseName is specified by the client application
	databaseName string
}

func (s *SingleUseConnector) abort(err error) {
	if s.clientConn == nil {
		return
	}

	pgError, ok := err.(*protocol.Error)
	// Not an expected protocol error, so we wrap it in a protocol error, so
	// the client can see what went wrong.
	if !ok {
		pgError = &protocol.Error{
			Severity: protocol.ErrorSeverityFatal,
			Code:     protocol.ErrorCodeInternalError,
			Message:  err.Error(),
		}
	}

	s.clientConn.Write(pgError.GetPacket())
}

func (s *SingleUseConnector) sslNotSupported() {
	if s.clientConn == nil {
		return
	}
	s.clientConn.Write([]byte{protocol.SSLNotAllowed})
}

// Connect implements the tcp.Connector func signature.
//
// It is the main method of the SingleUseConnector. It:
//   1. Constructs connection details from the provided credentials map.
//   2. Dials the backend using credentials.
//   3. Runs through the connection phase steps to authenticate.
//
// Connect requires "host", "port", "username" and "password" credentials.
func (s *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	var err error

	s.clientConn = clientConn

	if err = s.Startup(); err != nil {
		// connection closed by the client (f.e. SSLMode=require). Do not send error to the closed connection.
		if err != io.EOF {
			s.abort(err)
		}
		return nil, err
	}

	if len(credentialValuesByID["address"]) > 0 {
		s.logger.Warnf("'address' has been deprecated for PG connector. " +
			"Please use 'host' and 'port' instead.'")
	}

	var credIDs []string
	for credID := range credentialValuesByID {
		credIDs = append(credIDs, credID)
	}
	s.logger.Debugf("backend connection parameters: %s", credIDs)

	s.connectionDetails, err = NewConnectionDetails(credentialValuesByID)
	if err != nil {
		s.abort(err)
		return nil, err
	}

	if err = s.ConnectToBackend(); err != nil {
		s.abort(err)
		return nil, err
	}

	return s.backendConn, nil
}
