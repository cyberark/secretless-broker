package mssql

import (
	"context"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"

	mssql "github.com/denisenkom/go-mssqldb"
)

// SingleUseConnector is used to create an authenticated connection to an MSSQL target
type SingleUseConnector struct {
	backendConn net.Conn
	clientConn  net.Conn
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

	// 1. Customize PreLogin Handshake to not include ecnryption
	err := connector.customizePreLoginRequest()
	if err != nil {
		connector.logger.Errorf("Failed to handle client prelogin request: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// 2. Prepare connection details formatted for MsSQL
	connDetails, err := NewConnectionDetails(credentialValuesByID)
	if err != nil {
		connector.sendErrorToClient()
		return nil, err
	}

	// 3. Create a new MsSQL connector
	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := mssql.NewConnector(dataSourceName(connDetails))
	if err != nil {
		connector.logger.Errorf("Failed to create a go-mssqldb connector: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// 4. Kick off authentication through our third party connector
	driverConn, err := driverConnector.Connect(context.Background())
	if err != nil {
		connector.logger.Errorf("failed to connect to mssql server: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// TODO: 	5.	Send the prelogin response to the client (#1014)
	// TODO: 	6.	Send the login response to the client 	 (#1016)
	// TODO: 	Alt. Send an error to the client and return nil. (#1013)

	// TODO: Replace this with an actual 'ok' message from the server
	if _, err = clientConn.Write(connector.CreateAuthenticationOKMessage()); err != nil {
		connector.logger.Errorf("Failed to send a successful authentication"+
			" response to the client"+
			": %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// Verify the driverConn is an mssql driverConn object and get its underlying transport
	mssqlConn := driverConn.(*mssql.Conn)
	connector.backendConn = mssqlConn.NetConn()

	return connector.backendConn, nil
}

func (connector *SingleUseConnector) customizePreLoginRequest() error {
	// using the default packet size of 4096
	// (see https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-network-packet-size-server-configuration-option)
	clientBuffer := mssql.NewTdsBuffer(4096, connector.clientConn)
	preloginRequest, err := mssql.ReadPreloginWithPacketType(clientBuffer, mssql.PackPrelogin)
	if err != nil {
		return fmt.Errorf("failed to read prelogin request: %s", err)
	}

	// According to the mssql docs, The client can use the VERSION returned from
	// the server to determine which features SHOULD be enabled or disabled.
	// TODO: Extract version from server instead of hard-coded one.
	// we use now the version of the sql server in the test
	preloginRequest[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}

	// Remove Client SSL Capability from Server Handshake Packet
	// to force client to connect to Secretless without SSL
	preloginRequest[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

	// According to the docs, this value should be empty when being sent from
	// the server to the client.
	preloginRequest[mssql.PreloginTHREADID] = []byte{}

	err = mssql.WritePreloginWithPacketType(clientBuffer, preloginRequest, mssql.PackReply)
	if err != nil {
		return fmt.Errorf("failed to write prelogin response: %s", err)
	}

	// we actually don't need the client's handshake response.
	// we just need for them to not be blocked
	err = clientBuffer.ReadNextPacket()
	if err != nil {
		return fmt.Errorf("failed to read client login message: %s", err)
	}

	return nil
}

// TODO: Add ability to receive an MSSQL error and send it to the client (#1013)
func (connector *SingleUseConnector) sendErrorToClient() {
	mssqlError := connector.CreateGenericErrorMessage()
	if _, e := connector.clientConn.Write(mssqlError); e != nil {
		connector.logger.Errorf("failed to write error %s to MSSQL client", e)
	}
}

func dataSourceName(connDetails *ConnectionDetails) string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s",
		connDetails.Username,
		connDetails.Password,
		connDetails.Address(),
	)
}
