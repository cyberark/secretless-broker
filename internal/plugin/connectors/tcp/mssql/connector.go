package mssql

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
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
	newTdsBuffer      types.NewTdsBufferFunc
}

// NewSingleUseConnector creates a new SingleUseConnector
func NewSingleUseConnector(logger log.Logger) *SingleUseConnector {
	return NewSingleUseConnectorWithOptions(
		logger,
		NewMSSQLConnector,
		ReadPreloginWithPacketType,
		WritePreloginWithPacketType,
		NewTdsBuffer,
	)
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

// ReadPreloginWithPacketType is the production version of our readPrelogin
// dependency, which delegates to the actual 3rd party driver.
func ReadPreloginWithPacketType(
	rawTdsBuffer interface{},
	rawPktType interface{},
) (map[uint8][]byte, error) {
	// This can never fail unless mssql package changes: panicking is fine
	tdsBuffer := rawTdsBuffer.(*mssql.TdsBuffer)
	pktType := rawPktType.(mssql.PacketType)
	return mssql.ReadPreloginWithPacketType(tdsBuffer, pktType)
}

// WritePreloginWithPacketType is the production version of our writePrelogin
// dependency, which delegates to the actual 3rd party driver.
func WritePreloginWithPacketType(
	rawTdsBuffer interface{},
	fields map[uint8][]byte,
	rawPktType interface{},
) error {
	// This can never fail unless mssql package changes: panicking is fine
	tdsBuffer := rawTdsBuffer.(*mssql.TdsBuffer)
	pktType := rawPktType.(mssql.PacketType)
	return mssql.WritePreloginWithPacketType(tdsBuffer, fields, pktType)
}

// https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-network-packet-size-server-configuration-option
// Default packet size remains at 4096 bytes
const bufferSize uint16 = 4096

func NewTdsBuffer(transport io.ReadWriteCloser) types.ReadNextPacketer {
	return mssql.NewTdsBuffer(bufferSize, transport)
}

// NewSingleUseConnectorWithOptions creates a new SingleUseConnector, and allows
// you to specify the newMSSQLConnector explicitly.  Intended to be used in unit
// tests only.
func NewSingleUseConnectorWithOptions(
	logger log.Logger,
	newMSSQLConnector types.NewMSSQLConnectorFunc,
	readPrelogin types.ReadPreloginFunc,
	writePrelogin types.WritePreloginFunc,
	newTdsBuffer types.NewTdsBufferFunc,
) *SingleUseConnector {
	return &SingleUseConnector{
		logger:            logger,
		newMSSQLConnector: newMSSQLConnector,
		readPrelogin:      readPrelogin,
		writePrelogin:     writePrelogin,
		newTdsBuffer:      newTdsBuffer,
	}
}

type connectResult struct {
	conn *mssql.Conn
	err  error
}

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
	clientBuffer := connector.newTdsBuffer(connector.clientConn)
	_, err := connector.readPrelogin(clientBuffer, mssql.PackPrelogin)
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		wrappedError := errors.Wrap(err, "failed to read prelogin request")
		connector.sendError(wrappedError)
		return nil, wrappedError
	}

	// Prepare connection details from the client, formatted for MsSQL
	connDetails, err := NewConnectionDetails(credentialValuesByID)
	if err != nil {
		wrappedError := errors.Wrap(err, "unable to create new connection details")
		connector.sendError(wrappedError)
		return nil, wrappedError
	}

	// Create a new MsSQL connector
	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := connector.newMSSQLConnector(dataSourceName(connDetails))
	if err != nil {
		wrappedError := errors.Wrap(err, "failed to create a go-mssqldb connector")
		connector.sendError(wrappedError)
		return nil, wrappedError
	}

	// Create the context for our connection
	ctx := context.Background()

	// Set a 'marker' for when the driver has finished connecting to the server
	connectionResultChan := make(chan connectResult)

	// Create a ConnectIntercept for exchanging values with the driver via context and
	// ensure it gets cleaned up when this method returns
	connInterceptor := mssql.NewConnectInterceptor()
	defer connInterceptor.Close()

	// Add connInterceptor to the context for exchanging values with the driver
	loginContext := mssql.NewContextWithConnectInterceptor(ctx, connInterceptor)

	// Build a new driver connection
	var netConner types.NetConner

	go func() {
		// Kick off authentication through our third party connector
		netConner, err = driverConnector.Connect(loginContext)

		connectionResultChan <- connectResult{
			conn: netConner.(*mssql.Conn),
			err:  err,
		}
	}()

	var loginAck mssql.LoginAckStruct
	// Block continuation until error or driver has completed connection
responseLoop:
	for {
		select {
		// transient states
		//
		// preLoginResponse is received from the server
		case preloginResponse := <-connInterceptor.PreLoginResponse:
			// Since the communication between the client and Secretless must be unencrypted,
			// we fool the client into thinking that it's talking to a server that does not support
			// encryption.
			preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

			// Write the prelogin packet back to the user
			err = connector.writePrelogin(connector.clientBuff, preloginResponse)
			if err != nil {
				wrappedError := errors.Wrap(err, "failed to write prelogin response")
				connector.sendError(wrappedError)
				return nil, wrappedError
			}

			// We parse the client's Login packet so that we can pass on params to the server.
			clientLogin, err := mssql.ReadLogin(connector.clientBuff)
			if err == io.EOF {
				return nil, err
			}
			if err != nil {
				wrappedError := errors.Wrap(err, "failed to handle client login")
				connector.sendError(wrappedError)
				return nil, wrappedError
			}

			// Send the client login to the mssql driver which it can use to pass client params
			// to the server. This channel is set as a value on the context passed to the mssql
			// driver on construction.
			connInterceptor.ClientLogin <- *clientLogin
			// TODO: where this is consumed inside go-mssqldb, it real
			break
		// a loginAck is sent from the server, this will be followed by connectionResultChan
		// unless there's an issue
		case loginAck = <-connInterceptor.ServerLoginAck:
			break

			// terminating states
			//
		// a protocol error is received from the server
		case protocolErr := <-connInterceptor.ServerError:
			err = protocolErr
			break responseLoop
		// connect is finished. this case is visited when there is non-protocol error or
		// after loginAck
		case res := <-connectionResultChan:
			if res.err != nil {
				err = res.err
			} else {
				// Verify the driverConn is an mssql driverConn object and get its
				// underlying transport
				connector.backendConn = res.conn.NetConn()
			}

			break responseLoop
		}
	}

	if err != nil {
		connector.sendError(err)
		return nil, err
	}

	if err = mssql.WriteLoginAck(connector.clientBuff, loginAck); err != nil {
		wrappedError := errors.Wrap(
			err,
			"failed to send a successful authentication response to client",
		)
		connector.sendError(wrappedError)
		return nil, wrappedError
	}

	return connector.backendConn, nil
}

func (connector *SingleUseConnector) sendError(err error) {
	connector.logger.Error(err)
	var protocolErr mssql.Error
	if _protocolErr, ok := err.(mssql.Error); ok {
		protocolErr = _protocolErr
	} else {
		protocolErr = mssql.Error{
			// SQL Error Number - currently using 18456 (login failed for user)
			// TODO: Find generic error number
			Number: 18456,
			// State -
			// TODO: better understand this.
			State: 0x01,
			// Severity Class - 16 indicates a general error that can be corrected by the user.
			Class:      16,
			Message:    errors.Wrap(err, "secretless").Error(),
			ServerName: "secretless",
			ProcName:   "",
			LineNo:     0,
		}
	}
	_ = mssql.WriteError72(connector.clientBuff, protocolErr)
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
