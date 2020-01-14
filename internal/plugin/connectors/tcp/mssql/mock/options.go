package mock

import (
	"context"
	"net"

	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	mssql "github.com/denisenkom/go-mssqldb"
)

// MSSQLConnectorCtorOptions represents options when creating a MSSQLConnectorCtor. This
// makes it possible to customize some key values in a Connect double:
// func Connect(...) (net.Conn, error)
type MSSQLConnectorCtorOptions struct {
	// Backend connection returned from the call to Connect
	BackendConn net.Conn
	// Error returned from the call to Connect
	Err error
	// PreloginResponse from the server passed to interceptor during Connect
	ServerPreloginResponse map[uint8][]byte
	// Pointer to unload client login request to, taken from interceptor during Connect
	ClientLoginRequestPtr **mssql.LoginRequest
	// LoginResponse from the server passed to interceptor during Connect
	ServerLoginResponse *mssql.LoginResponse
}

// MSSQLConnectorCtorOption is the 'functional option' complement to
// MSSQLConnectorCtorOptions.
type MSSQLConnectorCtorOption func(*MSSQLConnectorCtorOptions)

// MSSQLConnectorCtor is a 'functional option' of types.ConnectorOption. It generates
// a customizable MSSQLConnectorCtor using MSSQLConnectorCtorOptions, then sets it on the
// ConnectorOption passed to it.
//
// Empty values of MSSQLConnectorCtorOption result default:
// ServerPreloginResponse => map[uint8][]byte{}
// ServerLoginResponse => &mssql.LoginResponse{}
// BackendConn => nil
// BackendConn => nil
// ClientLoginRequestPtr => nil (this means ClientLoginRequest is read from the channel to
// a vacuum)
func MSSQLConnectorCtor(setters ...MSSQLConnectorCtorOption) types.ConnectorOption {
	args := &MSSQLConnectorCtorOptions{}

	for _, setter := range setters {
		setter(args)
	}

	return func(connectorArgs *types.ConnectorOptions) {
		connectorArgs.NewMSSQLConnector = NewSuccessfulMSSQLConnectorCtor(
			func(ctx context.Context) (net.Conn, error) {
				interceptor := mssql.ConnectInterceptorFromContext(ctx)

				if args.ServerPreloginResponse == nil {
					args.ServerPreloginResponse = map[uint8][]byte{}
				}
				interceptor.ServerPreLoginResponse <- args.ServerPreloginResponse

				if args.ClientLoginRequestPtr != nil {
					*args.ClientLoginRequestPtr = <-interceptor.ClientLoginRequest
				} else {
					<-interceptor.ClientLoginRequest
				}

				if args.ServerLoginResponse == nil {
					args.ServerLoginResponse = &mssql.LoginResponse{}
				}
				interceptor.ServerLoginResponse <- args.ServerLoginResponse

				return args.BackendConn, args.Err
			},
		)
	}
}

// DefaultConnectorOptions returns a setter that will set ConnectorOptions to
// mocks that result in success.
func DefaultConnectorOptions() types.ConnectorOption {
	return func(connectOptions *types.ConnectorOptions) {
		connectOptions.Logger = logmock.NewLogger()
		connectOptions.NewMSSQLConnector = NewSuccessfulMSSQLConnectorCtor(
			func(ctx context.Context) (net.Conn, error) {
				interceptor := mssql.ConnectInterceptorFromContext(ctx)

				interceptor.ServerPreLoginResponse <- map[uint8][]byte{}

				<-interceptor.ClientLoginRequest

				interceptor.ServerLoginResponse <- &mssql.LoginResponse{}

				return NewNetConn(nil), nil
			},
		)
		connectOptions.ReadPreloginRequest = SuccessfulReadPreloginRequest
		connectOptions.WritePreloginResponse = SuccessfulWritePreloginResponse
		connectOptions.ReadLoginRequest = SuccessfulReadLoginRequest
		connectOptions.WriteLoginResponse = SuccessfulWriteLoginResponse
		connectOptions.WriteError = SuccessfulWriteError
		connectOptions.NewTdsBuffer = FakeTdsBufferCtor
	}
}
