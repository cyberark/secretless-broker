package mssql

import (
	"context"
	"fmt"
	"net"
	"database/sql/driver"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/third_party/ctxtypes"

	mssql "github.com/denisenkom/go-mssqldb"
)

// SingleUseConnector is used to create an authenticated connection to an MSSQL target
type SingleUseConnector struct {
	backendConn net.Conn
	clientConn  net.Conn
	logger      log.Logger
}

const bufferSize uint16 = 4086

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

	// Read the prelogin request from the client, but we don't need to do anything
	// we just need to make sure the client isn't blocked
	clientBuffer := mssql.NewTdsBuffer(bufferSize, connector.clientConn)
	_, err := mssql.ReadPreloginWithPacketType(clientBuffer, mssql.PackPrelogin)
	if err != nil {
		return nil, fmt.Errorf("failed to read prelogin request: %s", err)
	}

	// 2. Prepare connection details from the client, formatted for MsSQL
	connDetails, err := NewConnectionDetails(credentialValuesByID)
	if err != nil {
		connector.sendErrorToClient()
		return nil, err
	}

	// 3. Create a new MsSQL connector
	//    Using DSN (Data Source Name) string because gomssql forces us to.
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

	// Create the context with a channel we will use
	parentContext := context.Background()
	ch := make(chan map[uint8][]byte)

	// Set a 'marker' for when the driver has finished connecting
	driverConnSet := make(chan struct {})

	// Add channels to the context, to retrive information from the driver
	childContext := context.WithValue(parentContext, ctxtypes.PreLoginResponseKey, ch)

	// Build a new driver connection
	var driverConn driver.Conn

	go func() {
		// Kick off authentication through our third party connector
		driverConn, err = driverConnector.Connect(childContext)
		driverConnSet <- struct{}{}

		if err != nil {
			connector.logger.Errorf("failed to connect to mssql server: %s", err)
			connector.sendErrorToClient()
		}
	} ()

	// Blocks continuation until we've received the preLoginResponse from the driver
	preloginResponse := <- ch
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

	// Write the prelogin packet back to the user
	err = mssql.WritePreloginWithPacketType(clientBuffer, preloginResponse, mssql.PackReply)
	if err != nil {
		return nil, fmt.Errorf("failed to write prelogin response: %s", err)
	}

	// We actually don't need the client's handshake response;
	// we just need for them to not be blocked
	err = clientBuffer.ReadNextPacket()
	if err != nil {
		connector.logger.Errorf("Failed to handle client prelogin request: %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// TODO: 	5.	Send the prelogin response to the client (#1014)
	// TODO: 	6.	Send the login response to the client 	 (#1016)
	// TODO: 	Alt. Send an error to the client and return nil. (#1013)

	// TODO: 	Replace this with an actual 'ok' message from the server
	if _, err = clientConn.Write(connector.CreateAuthenticationOKMessage()); err != nil {
		connector.logger.Errorf("Failed to send a successful authentication"+
			" response to the client"+
			": %s", err)
		connector.sendErrorToClient()
		return nil, err
	}

	// Block continuation until driver has completed connection
	<- driverConnSet

	// Verify the driverConn is an mssql driverConn object and get its underlying transport
	mssqlConn := driverConn.(*mssql.Conn)
	connector.backendConn = mssqlConn.NetConn()

	return connector.backendConn, nil
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
