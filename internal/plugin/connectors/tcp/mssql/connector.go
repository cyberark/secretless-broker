package mssql

import (
	"context"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/third_party/ctxtypes"

	mssql "github.com/denisenkom/go-mssqldb"
)

/*

This connector acts as a layer between the client and the driver, where the
driver handles communication with the server itself. As such, some stages are
handled independently of secretless, while other stages require the interception
or modification of the respective requests or responses.

Overview of the connection process

+---------+              +-------------+                      +---------+               +-------+
| Client  |              | Secretless  |                      | Driver  |               | MsSQL |
+---------+              +-------------+                      +---------+               +-------+
     |                          |                                  |                        |
     | Prelogin Request         |                                  |                        |
     |------------------------->|                                  |                        |
     |                          |                                  |                        |
     |                          | Connect(Context)                 |                        |
     |                          |--------------------------------->|                        |
     |                          | ----------------------------\    |                        |
     |                          |-| Context contains channels |    |                        |
     |                          | | for intercepting data     |    |                        |
     |                          | |---------------------------|    |                        |
     |                          |                                  | Prelogin Request       |
     |                          |                                  |----------------------->|
     |                          |                                  |                        |
     |                          |                                  |      Prelogin Response |
     |                          |                                  |<-----------------------|
     |                          |                                  |                        |
     |                          |      Intercept Prelogin Response |                        |
     |                          |<---------------------------------|                        |
     |                          |                                  |                        |
     |                          | Modify Prelogin Response         |                        |
     |                          |------------------------          |                        |
     |                          |                       |          |                        |
     |                          |<-----------------------          |                        |
     |                          |                                  |                        |
     |        Prelogin Response |                                  |                        |
     |<-------------------------|                                  |                        |
     |                          |                                  |                        |
     |                          |                                  | Handshake Request      |
     |                          |                                  |----------------------->|
     |                          |                                  |                        |
     |                          |                                  |     Handshake Response |
     |                          |                                  |<-----------------------|
     |                          |                                  |                        |
     |                          |                                  | Login Request          |
     |                          |                                  |----------------------->|
     |                          |                                  |                        |
     |                          |                                  |         Login Response |
     |                          |                                  |<-----------------------|
     |                          |                                  |                        |
     |                          |         Login success or failure |                        |
     |                          |<---------------------------------|                        |
     |                          |                                  |                        |
     |    Login Response (Fake) |                                  |                        |
     |<-------------------------|                                  |                        |
     |                          |                                  |                        |

	Note: The above diagram was created using https://textart.io/sequence and the
	following source:

	object Client Secretless Driver MsSQL
	Client->Secretless: Prelogin Request
	Secretless->Driver: Connect(context)
	note right of Secretless: Context contains channels \n for intercepting data
	Driver->MsSQL: Prelogin Request
	MsSQL->Driver: Prelogin Response
	MsSQL->Secretless: Intercept Prelogin Response
	Secretless->Secretless: Modify Prelogin Response
	Secretless->Client: Prelogin Response
	Driver->MsSQL:Handshake Request
	MsSQL->Driver: Handshake Response
	Driver->MsSQL: Login Request
	MsSQL->Driver: Login Response
	Driver->Secretless: Login success or failure
	Secretless->Client: Login Response (Premade)
*/

// SingleUseConnector is used to create an authenticated connection to an MSSQL target
type SingleUseConnector struct {
	backendConn       net.Conn
	clientConn        net.Conn
	logger            log.Logger
	newMSSQLConnector types.NewMSSQLConnectorFunc
	readPrelogin      types.ReadPreloginFunc
	writePrelogin     types.WritePreloginFunc
}

// NewMSSQLConnector is the production implementation of NewMSSQLConnectorFunc,
// used for creating mssql.Connector instances.  We need to wrap the raw
// constructor provided by mssql (ie, mssql.NewConnector) in this function so
// that it returns an interface, which enables us to mock it in unit tests.
func NewMSSQLConnector(dsn string) (types.MSSQLConnector, error) {
	connector, err := mssql.NewConnector(dsn)
	fn := func(ctx context.Context) (types.NetConner, error) {
		driverConn, err := connector.Connect(ctx)
		if err != nil {
			return nil, err
		}
		// This can never fail unless mssql package changes: panicking is fine
		mssqlConn := driverConn.(types.NetConner)
		return mssqlConn, nil
	}
	return types.MSSQLConnectorFunc(fn), err
}

// NewSingleUseConnector creates a new SingleUseConnector
func NewSingleUseConnector(logger log.Logger) *SingleUseConnector {
	return NewSingleUseConnectorWithOptions(
		logger,
		NewMSSQLConnector,
		mssql.ReadPreloginWithPacketType,
		mssql.WritePreloginWithPacketType,
	)
}

// NewSingleUseConnector creates a new SingleUseConnector, and allows you to
// specify the newMSSQLConnector explicitly.  Intended to be used in unit tests
// only.
func NewSingleUseConnectorWithOptions(
	logger log.Logger,
	newMSSQLConnector types.NewMSSQLConnectorFunc,
	readPrelogin types.ReadPreloginFunc,
	writePrelogin types.WritePreloginFunc,
) *SingleUseConnector {
	return &SingleUseConnector{
		logger:            logger,
		newMSSQLConnector: newMSSQLConnector,
		readPrelogin:      readPrelogin,
		writePrelogin:     writePrelogin,
	}
}

// https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-network-packet-size-server-configuration-option
// Default packet size remains at 4096 bytes
const bufferSize uint16 = 4096

// Connect implements the tcp.Connector func signature
//
// It is the main method of the SingleUseConnector. It:
//   1. Constructs connection details from the provided credentials map.
//   2. Reads the client PreLogin request
//   3. Adds channels to the context that can intercept data from the driver
//   4. Initiates authentication and connection to MsSQL through the third-party driver
//   5. Intercepts PreLogin response from the driver
//	 6. Customizes the PreLogin response to meet Secretless standards
//		and sends it to the user
//
// Connect requires "host", "port", "username" and "password" credentials.
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {

	connector.clientConn = clientConn

	// Secretless _is_ the client with respect to the server, and there is
	// nothing in the pre-login handshake that needs to be passed along.
	// Secretless simply reads it from the client and throws it away, so that
	// client can advance to the next stage of the process.  Otherwise the
	// client would block forever waiting for its pre-login handshake to be
	// read.
	clientBuffer := mssql.NewTdsBuffer(bufferSize, connector.clientConn)
	_, err := connector.readPrelogin(clientBuffer, mssql.PackPrelogin)
	if err != nil {
		return connector.sendError("failed to read prelogin request: %s", err)
	}

	// Prepare connection details from the client, formatted for MsSQL
	connDetails, err := NewConnectionDetails(credentialValuesByID)
	if err != nil {
		return connector.sendError("Unable to create new connection details: %s", err)
	}

	// Create a new MsSQL connector
	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := connector.newMSSQLConnector(dataSourceName(connDetails))
	if err != nil {
		return connector.sendError("Failed to create a go-mssqldb connector: %s", err)
	}

	// Create the context for our connection
	ctx := context.Background()

	// Create a channel for receiving the prelogin response through the context
	preLoginResponseChannel := make(chan map[uint8][]byte)

	// Set a 'marker' for when the driver has finished connecting to the server
	connectPhaseFinished := make(chan struct{})

	// Add channels to the context, to retrive information from the driver
	loginContext := context.WithValue(ctx, ctxtypes.PreLoginResponseKey,
		preLoginResponseChannel)

	// Build a new driver connection
	var netConner types.NetConner

	go func() {
		// Kick off authentication through our third party connector
		netConner, err = driverConnector.Connect(loginContext)
		connectPhaseFinished <- struct{}{}

		if err != nil {
			_, err = connector.sendError("failed to connect to mssql server: %s", err)
		}
	}()

	// Blocks continuation until we've received the preLoginResponse from the driver
	preloginResponse := <-preLoginResponseChannel

	// Since the communication between the client and Secretless must be unencrypted,
	// we fool the client into thinking that it's talking to a server that does not support
	// encryption.
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

	// Write the prelogin packet back to the user
	err = connector.writePrelogin(clientBuffer, preloginResponse, mssql.PackReply)
	if err != nil {
		return connector.sendError("failed to write prelogin response: %s", err)
	}

	// Just like above, we don't need to concern ourselves with what the client would
	// typically be sending to the server, since we are handling all communication with
	// the server ourselves. We read the next packet so the client isn't blocked,
	// while the driver handles the handshake.
	err = clientBuffer.ReadNextPacket()
	if err != nil {
		return connector.sendError("Failed to handle client prelogin request: %s", err)
	}

	// TODO: 	Send the login response to the client 	 (#1016)
	// TODO: 	Verify appropriate errors are passed to the client (#1013)

	// TODO: 	Replace this with an actual 'ok' message from the server
	//			once login completes within the driver
	// TODO:    Rename this to "AuthenticationOKMessage"
	if _, err = clientConn.Write(connector.CreateAuthenticationOKMessage()); err != nil {
		return connector.sendError("Failed to send a successful authentication"+
			" response to the client"+
			": %s", err)
	}

	// Block continuation until driver has completed connection
	<-connectPhaseFinished

	connector.backendConn = netConner.NetConn()

	return connector.backendConn, nil
}

// TODO: Add ability to receive an MSSQL error and send it to the client (#1013)
func (connector *SingleUseConnector) sendError(message string, err error) (net.Conn,
	error) {
	connector.logger.Errorf(message, err)
	connector.sendErrorToClient()
	return nil, err
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
