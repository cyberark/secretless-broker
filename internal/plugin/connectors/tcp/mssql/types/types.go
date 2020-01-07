package types

import (
	"context"
	"io"
	"net"
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

// ReadPreloginFunc defines the type of the func that reads the prelogin packet.
// The production version is implemented by mssql.ReadPreloginWithPacketType.
type ReadPreloginFunc func(
	tdsBuffer io.ReadWriteCloser,
	pktType uint8) (map[uint8][]byte, error)

// WritePreloginFunc defines the type of the func that writes the prelogin
// packet. The production version is implemented by
// mssql.WritePreloginWithPacketType.
type WritePreloginFunc func(
	tdsBuffer io.ReadWriteCloser,
	fields map[uint8][]byte,
	pktType uint8) error

// ReadLoginFunc defins the type of the func that reads the client's login
// packet.  The production version is implemented by:
//     mssql.ReadLogin(r *TdsBuffer) (*Login, error)
// Note that, in order to avoid a concrete dependency on mssql in this package,
// we must replace TdsBuffer with ReadNextPacketer and *Login with interface{}.
// That interface{} will then be case back to a *Login by the receiving code
// inside the driver package.
type ReadLoginFunc func(r io.ReadWriteCloser) (interface{}, error)

// TdsBufferCtor represents the constructor of a TdsBuffer, in a form
// suitable for unit tests.
//
// Note the signature does not mention TdsBuffers.  This is because our code is
// only concerned with the ReadWriteCloser closer functionality, and our doubles
// can be ReadWriteClosers.  The production implementation needs of course to
// return a real TdsBuffer (which _is_ a ReadWriteCloser), and so we've chosen a
// name that reflects the production purpose.
type TdsBufferCtor func(transport io.ReadWriteCloser) io.ReadWriteCloser
