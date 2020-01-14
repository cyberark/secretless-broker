package mssql

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/pkg/errors"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	mssql "github.com/denisenkom/go-mssqldb"
)

/*

This connector acts as a layer between the client and the driver, where the
driver handles communication with the server itself. As such, some stages are
handled independently of secretless, while other stages require the interception
or modification of the respective requests or responses.

Overview of the connection process

+---------+              +-------------+                      +---------+               +-------+
| Client  |              | Secretless  |                      | Driver  |               | MSSQL |
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
     |                          |         Intercept Login Response |                        |
     |                          |<---------------------------------|                        |
     |                          |                                  |                        |
     |                          |      Backend Connection or error |                        |
     |                          |<---------------------------------|                        |
     |                          |                                  |                        |
     |  Login Response or error |                                  |                        |
     |<-------------------------|                                  |                        |
     |                          |                                  |                        |

	Note: The above diagram was created using https://textart.io/sequence and the
	following source:

	object Client Secretless Driver MSSQL
	Client->Secretless: Prelogin Request
	Secretless->Driver: Connect(context)
	note right of Secretless: Context contains channels \n for intercepting data
	Driver->MSSQL: Prelogin Request
	MSSQL->Driver: Prelogin Response
	MSSQL->Secretless: Intercept Prelogin Response
	Secretless->Secretless: Modify Prelogin Response
	Secretless->Client: Prelogin Response
	Driver->MSSQL:Handshake Request
	MSSQL->Driver: Handshake Response
	Driver->MSSQL: Login Request
	MSSQL->Driver: Login Response
	Driver->Secretless: Intercept Login Response
	Driver->Secretless: Backend Connection or error
	Secretless->Client: Login Response or error
*/

// SingleUseConnector is used to create an authenticated connection to an MSSQL target
type SingleUseConnector struct {
	clientConn net.Conn
	clientBuff io.ReadWriteCloser
	logger     log.Logger
	// Note: We're following standard ctor naming practices with this field.
	newMSSQLConnector     types.MSSQLConnectorCtor
	readPreloginRequest   types.ReadPreloginRequestFunc
	writePreloginResponse types.WritePreloginResponseFunc
	readLoginRequest      types.ReadLoginRequestFunc
	writeLoginResponse    types.WriteLoginResponseFunc
	writeError            types.WriteErrorFunc
	newTdsBuffer          types.TdsBufferCtor
}

// NewSingleUseConnector creates a new production SingleUseConnector.
// This uses the production version of the dependencies, which delegate to the actual 3rd
// party driver.
func NewSingleUseConnector(logger log.Logger) *SingleUseConnector {
	return NewSingleUseConnectorWithOptions(func(opt *types.ConnectorOptions) {
		opt.Logger = logger
		opt.NewMSSQLConnector = NewMSSQLConnector
		opt.ReadPreloginRequest = mssql.ReadPreloginRequest
		opt.WritePreloginResponse = mssql.WritePreloginResponse
		opt.ReadLoginRequest = mssql.ReadLoginRequest
		opt.WriteLoginResponse = mssql.WriteLoginResponse
		opt.WriteError = mssql.WriteError72
		// NewIdempotentDefaultTdsBuffer is wrapped so that it conforms to the
		// types.TdsBufferCtor func signature
		opt.NewTdsBuffer = func(transport io.ReadWriteCloser) io.ReadWriteCloser {
			return mssql.NewIdempotentDefaultTdsBuffer(transport)
		}
	})
}

// NewMSSQLConnector is the production implementation of MSSQLConnectorCtor,
// used for creating mssql.Connector instances.  We need to wrap the raw
// constructor provided by mssql (ie, mssql.NewConnector) in this function so
// that it returns an interface, which enables us to mock it in unit tests.
func NewMSSQLConnector(dsn string) (types.MSSQLConnector, error) {
	c, err := mssql.NewConnector(dsn)
	fn := func(ctx context.Context) (net.Conn, error) {
		driverConn, err := c.Connect(ctx)
		if err != nil {
			return nil, err
		}
		// This can never fail unless mssql package changes: panicking is fine
		mssqlConn := driverConn.(*mssql.Conn).NetConn()
		return mssqlConn, nil
	}
	return types.MSSQLConnectorFunc(fn), err
}

// NewSingleUseConnectorWithOptions creates a new SingleUseConnector, and allows
// you to specify the newMSSQLConnector explicitly.  Intended to be used in unit
// tests only.
func NewSingleUseConnectorWithOptions(
	options ...types.ConnectorOption,
) *SingleUseConnector {
	// Default Options
	args := &types.ConnectorOptions{}

	for _, option := range options {
		option(args)
	}

	return &SingleUseConnector{
		logger:                args.Logger,
		newMSSQLConnector:     args.NewMSSQLConnector,
		readPreloginRequest:   args.ReadPreloginRequest,
		writePreloginResponse: args.WritePreloginResponse,
		readLoginRequest:      args.ReadLoginRequest,
		writeLoginResponse:    args.WriteLoginResponse,
		writeError:            args.WriteError,
		newTdsBuffer:          args.NewTdsBuffer,
	}
}

type connectResult struct {
	conn net.Conn
	err  error
}

// Connect implements the tcp.Connector func signature
//
// It is the main method of the SingleUseConnector. It:
//   1. Reads the client PreLogin request
//   2. Constructs connection details from the provided credentials map
//   3. Adds a ConnectInterceptor to the context to exchange data with the driver via
//   	channels
//   4. Initiates authentication and connection to MSSQL through the third-party driver
//   5. Injects client's PreLogin request to the driver, which the driver incorporates
//   	into its PreLogin request to server
//   6. Intercepts PreLogin response or error from the driver
//	 7. Customizes the PreLogin response to meet Secretless standards and sends it to
//		client
//   8. Intercepts Login response or error from the driver, and simultaneously extracts
//   	net.Conn to server from the driver
//   9. Sends Login response or error to client
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
	connector.clientBuff = connector.newTdsBuffer(connector.clientConn)
	_, err := connector.readPreloginRequest(connector.clientBuff)
	if err != nil {
		wrappedError := errors.Wrap(err, "failed to read prelogin request from client")
		connector.writeErrorToClient(wrappedError)
		return nil, wrappedError
	}

	// Prepare connection details from the client, formatted for MSSQL
	// TODO: find out if it is possible to send errors during prelogin-phase
	// TODO: send error to client on failed credential validation
	connDetails := NewConnectionDetails(credentialValuesByID)

	// Create a new MSSQL connector
	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := connector.newMSSQLConnector(dataSourceName(connDetails))
	if err != nil {
		wrappedError := errors.Wrap(err, "failed to create a go-mssqldb connector")
		connector.writeErrorToClient(wrappedError)
		return nil, wrappedError
	}

	// Set a 'marker' for when the driver has finished connecting to the server
	connectionResultChan := make(chan connectResult)

	// Create a ConnectInterceptor for exchanging values with the driver via context
	connInterceptor := mssql.NewConnectInterceptor()

	// Add connInterceptor to the context for exchanging values with the driver
	loginContext := mssql.NewContextWithConnectInterceptor(context.Background(), connInterceptor)

	go func() {
		// Kick off authentication through our third party connector
		driverConn, err := driverConnector.Connect(loginContext)

		connectionResultChan <- connectResult{
			conn: driverConn,
			err:  err,
		}
	}()

	loginResp, backendConn, err := connector.waitForServerConnection(
		connInterceptor,
		connectionResultChan)

	if err != nil {
		connector.writeErrorToClient(err)
		return nil, err
	}

	if err = connector.writeLoginResponse(connector.clientBuff, loginResp); err != nil {
		wrappedError := errors.Wrap(
			err,
			"failed to send a successful authentication response to client",
		)
		connector.writeErrorToClient(wrappedError)
		return nil, wrappedError
	}

	return backendConn, nil
}

func protocolError(err error) mssql.Error {
	if _protocolErr, ok := errors.Cause(err).(mssql.Error); ok {
		return _protocolErr
	}

	return mssql.Error{
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

func (connector *SingleUseConnector) waitForServerConnection(
	interceptor *mssql.ConnectInterceptor,
	connResChan chan connectResult,
) (*mssql.LoginResponse, net.Conn, error) {
	var loginResponse *mssql.LoginResponse
	var err error

	// SECTION: wait for prelogin response
	//
	select {
	// preloginResponse is received from the server
	case preloginResponse := <-interceptor.ServerPreLoginResponse:
		if preloginResponse == nil {
			return nil, nil, errors.New("ServerPreLoginResponse is nil")
		}
		// Since the communication between the client and Secretless must be unencrypted,
		// we fool the client into thinking that it's talking to a server that does not
		// support encryption.
		preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

		// Write the prelogin packet back to the user
		err = connector.writePreloginResponse(connector.clientBuff, preloginResponse)
		if err != nil {
			wrappedError := errors.Wrap(
				err,
				"failed to write prelogin response to client")
			return nil, nil, wrappedError
		}

		// We parse the client's LoginRequest packet so that we can pass on params to the
		// server.
		clientLoginRequest, err := connector.readLoginRequest(connector.clientBuff)
		if err != nil {
			wrappedError := errors.Wrap(err, "failed to handle login from client")
			return nil, nil, wrappedError
		}

		// Send the client login to the mssql driver which it can use to pass client
		// params to the server. This channel is set as a value on the context passed to
		// the mssql driver on construction.
		interceptor.ClientLoginRequest <- clientLoginRequest
		break

	// error is received from connect
	case res := <-connResChan:
		if res.err == nil {
			panic("connect finished before preloginResponse without error")
		}
		return nil, nil, res.err
	}

	// SECTION: wait for server login ack
	//
	select {
	// loginResponse is received from the server
	case loginResponse = <-interceptor.ServerLoginResponse:
		if loginResponse == nil {
			return nil, nil, errors.New("ServerLoginResponse is nil")
		}
		break

	// error is received from connect
	case res := <-connResChan:
		if res.err == nil {
			panic("connect finished before loginResponse without error")
		}
		return nil, nil, res.err
	}

	// SECTION: wait for connection response
	//
	res := <-connResChan
	if res.err != nil {
		return nil, nil, res.err
	}

	return loginResponse, res.conn, nil
}

func (connector *SingleUseConnector) writeErrorToClient(err error) {
	_ = connector.writeError(connector.clientBuff, protocolError(err))
}

func dataSourceName(connDetails *ConnectionDetails) string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s",
		connDetails.Username,
		connDetails.Password,
		connDetails.Address(),
	)
}
