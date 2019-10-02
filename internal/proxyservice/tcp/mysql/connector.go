package mysql

import (
	"net"

	"github.com/cyberark/secretless-broker/internal/proxyservice/tcp/mysql/protocol"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
)

// Connector creates an authenticated connection to a MySQL target service.
//
type Connector struct {
	// The connections are decorated versions of net.Conn that allow us
	// to do read/writes according to the MySQL protocol.  Protocol level
	// details are thus encapsulated.  Within the MySQL code, we _only_
	// deal with these decorated versions.

	mySQLClientConn   *Connection
	mySQLBackendConn  *Connection
	connectionDetails *ConnectionDetails
	logger            log.Logger
}

// If the error is not already a MySQL protocol error, then wrap it in an
// "Unknown" MySQL protocol error, because the client understands only those.
//
func (connector *Connector) sendErrorToClient(err error) {
	mysqlErrorContainer, isProtocolErr := err.(protocol.ErrorContainer)
	if !isProtocolErr {
		mysqlErrorContainer = protocol.NewGenericError(err)
	}

	if e := connector.mySQLClientConn.write(mysqlErrorContainer.GetPacket()); e != nil {
		msg := "Attempted to write error %s to MySQL client but failed"
		connector.logger.Infof(msg, e)
	}
}

// Connect implements the tcp.Connector interface
//
// It is the main method of the Connector. It:
//   1. Constructs connection details from the provided credentials map.
//   1. Dials the backend using credentials.
//   3. Runs through the connection phase steps to authenticate.
//   4. Pipes all future bytes unaltered between client and server.
//
// Connect requires "host", "port", "username" and "password" credentials.
//
func (connector *Connector) Connect(
	clientConn net.Conn,
	credentialValuesByID plugin.CredentialValuesByID,
) (net.Conn, error) {

	// Upgrade to a decorated connection that handles protocol details for us
	// We need to do this first because sendErrorToClient uses this to send error messages.
	//
	connector.mySQLClientConn = NewClientConnection(clientConn)

	// 1. Construct connection details from the provided credentialValuesByID map.
	//
	connDetails, err := NewConnectionDetails(credentialValuesByID)
	if err != nil {
		connector.sendErrorToClient(err)
		return nil, err
	}

	// 2. Dials the backend.
	//
	rawBackendConn, err := net.Dial("tcp", connDetails.Address())
	if err != nil {
		connector.sendErrorToClient(err)
		return nil, err
	}

	connector.mySQLBackendConn = NewBackendConnection(rawBackendConn)

	// 3. Runs through the connection phase steps to authenticate.
	//
	connPhase := NewAuthenticationHandshake(
		connector.mySQLClientConn,
		connector.mySQLBackendConn,
		connDetails,
		)

	if err = connPhase.Run(); err != nil {
		connector.sendErrorToClient(err)
		return nil, err
	}

	backendConnection := connPhase.AuthenticatedBackendConn() // conn may have changed

	return backendConnection, nil
}
