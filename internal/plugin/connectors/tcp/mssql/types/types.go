package types

import (
	"context"
	"io"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	mssql "github.com/denisenkom/go-mssqldb"
)

// MSSQLConnectorCtor represents the constructor of an mssqlConnector. It
// exists so that its production implementation (mssql.NewConnector) can be
// swapped out in unit tests.  Note we keep MSSQL in the name to prevent
// confusion with the Secretless Connector.
type MSSQLConnectorCtor func(dsn string) (MSSQLConnector, error)

// MSSQLConnector captures the part of the 3rd party driver's mssql.Connector
// type that we care about -- its "Connect" method -- in an interface.  This
// allows us to mock that in our unit tests.
type MSSQLConnector interface {
	Connect(context.Context) (net.Conn, error)
}

// MSSQLConnectorFunc lets us treat a function (that matches the "Connect"
// signature) as an MSSQLConnector interface.
type MSSQLConnectorFunc func(context.Context) (net.Conn, error)

// Connect implements the MSSQLConnector interface.
func (fn MSSQLConnectorFunc) Connect(ctx context.Context) (net.Conn, error) {
	return fn(ctx)
}

// ReadPreloginRequestFunc defines the type of the func that reads the prelogin packet.
// The production version is implemented by mssql.ReadPreloginRequest.
type ReadPreloginRequestFunc func(
	r io.ReadWriteCloser,
) (map[uint8][]byte, error)

// WritePreloginResponseFunc defines the type of the func that writes the prelogin
// response packet. The production version is implemented by mssql.WritePreloginResponse.
type WritePreloginResponseFunc func(
	w io.ReadWriteCloser,
	fields map[uint8][]byte,
) error

// ReadLoginRequestFunc defines the type of the func that reads the client's login
// packet.  The production version is implemented by mssql.ReadLoginRequest.
type ReadLoginRequestFunc func(r io.ReadWriteCloser) (*mssql.LoginRequest, error)

// WriteErrorFunc defines the type of the func that writes an error packet. The production
// version is implemented by mssql.WriteError.
type WriteErrorFunc func(
	w io.ReadWriteCloser,
	err mssql.Error,
) error

// TdsBufferCtor represents the constructor of a TdsBuffer, in a form
// suitable for unit tests.
//
// Note the signature does not mention TdsBuffers.  This is because our code is
// only concerned with the ReadWriteCloser closer functionality, and our doubles
// can be ReadWriteClosers.  The production implementation needs of course to
// return a real TdsBuffer (which _is_ a ReadWriteCloser), and so we've chosen a
// name that reflects the production purpose.
type TdsBufferCtor func(transport io.ReadWriteCloser) io.ReadWriteCloser

// ConnectorOptions captures all the configuration options for a SingleUseConnector
type ConnectorOptions struct {
	Logger                log.Logger
	NewMSSQLConnector     MSSQLConnectorCtor
	ReadPreloginRequest   ReadPreloginRequestFunc
	WritePreloginResponse WritePreloginResponseFunc
	ReadLoginRequest      ReadLoginRequestFunc
	WriteError            WriteErrorFunc
	NewTdsBuffer          TdsBufferCtor
}

// ConnectorOption is the 'functional option' complement to ConnectorOptions
type ConnectorOption func(*ConnectorOptions)
