package mssql

import (
	"context"
	"net"
	"net/url"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"

	mssql "github.com/denisenkom/go-mssqldb"
)

// SingleUseConnector is used to create an authenticated connection to an MSSQL target
type SingleUseConnector struct {
	backendConn net.Conn
	clientConn  net.Conn
	clientLogin  *mssql.Login
	logger      log.Logger
}

// Connect implements the tcp.Connector func signature
//
// It is the main method of the SingleUseConnector. It:
//   1. Constructs connection details from the provided credentials map.
//   2. Dials the backend using credentials.
//   3. Runs through the connection phase steps to authenticate.
//
// Connect requires "host", "port", "username" and "password" credentials.
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {

	connector.clientConn = clientConn

	err := connector.performPreLoginHandshake()
	if err != nil {
		connector.logger.Errorf("Failed to handle client prelogin request: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	connDetails, err := NewConnectionDetails(credentialValuesByID, connector.clientLogin)
	if err != nil {
		connector.sendErrorToClient()
		return nil, err
	}

	err = connector.ConnectToBackend(dataSourceName(connDetails))
	if err != nil {
		connector.logger.Errorf("Failed to connect to backend: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// Now that the backend connection is authenticated, send the client
	// a successful authentication response.
	if _, err = clientConn.Write(connector.CreateAuthenticationOKMessage()); err != nil {
		connector.logger.Errorf("Failed to send a successful authentication"+
			" response to the client"+
			": %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	return connector.backendConn, nil
}

// ConnectToBackend establishes the connection to the target database and sets
// the backendConnection field.
func (connector *SingleUseConnector) ConnectToBackend(dataSourceName string) error {
	var err error

	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := mssql.NewConnector(dataSourceName)
	if err != nil {
		connector.logger.Errorf("Failed to create a go-mssqldb connector: %s", err)
		return err
	}

	driverConn, err := driverConnector.Connect(context.Background())
	if err != nil {
		connector.logger.Errorf("failed to connect to mssql server: %s", err)
		return err
	}

	// Verify the driverConn is an mssql driverConn object and get its underlying transport
	mssqlConn := driverConn.(*mssql.Conn)
	connector.backendConn = mssqlConn.NetConn()

	return nil
}

// TODO: add ability to receive an MSSQL error and send it to the client
func (connector *SingleUseConnector) sendErrorToClient() {
	mssqlError := connector.CreateGenericErrorMessage()
	if _, e := connector.clientConn.Write(mssqlError); e != nil {
		connector.logger.Errorf("failed to write error %s to MSSQL client", e)
	}
}

func dataSourceName(connDetails *ConnectionDetails) string {
	query := url.Values{}
	query.Add("app name", connDetails.AppName)
	query.Add("database", connDetails.Database)

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(connDetails.Username, connDetails.Password),
		Host:     connDetails.Address(),
		RawQuery: query.Encode(),
	}

	return u.String()
}
