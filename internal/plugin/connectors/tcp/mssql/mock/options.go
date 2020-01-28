package mock

import (
	"context"
	"net"

	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	mssql "github.com/denisenkom/go-mssqldb"
)

// MSSQLConnectorCtor is a mock which represents options when creating a
// mssql.MSSQLConnectorCtor. This makes it possible to customize some
// key values in a Connect double: func Connect(...) (net.Conn, error)
type MSSQLConnectorCtor struct {
	// Backend connection returned from the call to Connect
	BackendConn net.Conn
	// Error returned from the call to Connect
	Err error
	// PreloginResponse from the server passed to interceptor during Connect
	ServerPreloginResponse map[uint8][]byte
	// Pointer to unload client login request to, taken from interceptor during Connect
	ClientLoginRequestPtr **mssql.LoginRequest
}

// DefaultMSSQLConnectorCtor is the default constructor for MSSQLConnectorCtor
var DefaultMSSQLConnectorCtor = MSSQLConnectorCtor{
	BackendConn:            NewNetConn(nil),
	Err:                    nil,
	ServerPreloginResponse: map[uint8][]byte{},
	ClientLoginRequestPtr:  nil,
}

// DefaultConnectorOptions is a 'functional option' containing the default successful
// methods of each dependency
var DefaultConnectorOptions types.ConnectorOption = func(connectOptions *types.ConnectorOptions) {
	connectOptions.Logger = logmock.NewLogger()
	connectOptions.ReadPreloginRequest = SuccessfulReadPreloginRequest
	connectOptions.WritePreloginResponse = SuccessfulWritePreloginResponse
	connectOptions.ReadLoginRequest = SuccessfulReadLoginRequest
	connectOptions.WriteError = SuccessfulWriteError
	connectOptions.NewTdsBuffer = FakeTdsBufferCtor
	connectOptions.NewMSSQLConnector = DefaultNewMSSQLConnector
}

// DefaultNewMSSQLConnector returns an always successful MSSQLConnectorCtor
var DefaultNewMSSQLConnector = NewSuccessfulMSSQLConnectorCtor(
	func(ctx context.Context) (net.Conn, error) {
		interceptor := mssql.ConnectInterceptorFromContext(ctx)

		interceptor.ServerPreLoginResponse <- DefaultMSSQLConnectorCtor.ServerPreloginResponse

		<-interceptor.ClientLoginRequest

		return DefaultMSSQLConnectorCtor.BackendConn, DefaultMSSQLConnectorCtor.Err
	},
)

// CustomNewMSSQLConnectorOption allows us to inject a custom MSSQLConnectorCtor
// to control the output of the connect method
func CustomNewMSSQLConnectorOption(ctor MSSQLConnectorCtor) types.ConnectorOption {
	return func(options *types.ConnectorOptions) {
		options.NewMSSQLConnector = NewSuccessfulMSSQLConnectorCtor(
			func(ctx context.Context) (net.Conn, error) {
				interceptor := mssql.ConnectInterceptorFromContext(ctx)

				interceptor.ServerPreLoginResponse <- ctor.ServerPreloginResponse

				if ctor.ClientLoginRequestPtr != nil {
					*ctor.ClientLoginRequestPtr = <-interceptor.ClientLoginRequest
				} else {
					<-interceptor.ClientLoginRequest
				}

				return ctor.BackendConn, ctor.Err
			})
	}
}
