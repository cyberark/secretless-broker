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
	Connect(context.Context) (NetConner, error)
}

// NetConner is anything with a NetConn() method.  Ie, anything that can provide
// a net.Conn.  Note this rather silly name conforms to Go standard conventions
// for naming single method interfaces.
type NetConner interface {
	NetConn() net.Conn
}

// MSSQLConnectorFunc lets us treat a function (that matches the "Connect"
// signature) as an MSSQLConnector interface.
type MSSQLConnectorFunc func(context.Context) (NetConner, error)

// Connect implements the MSSQLConnector interface.
func (fn MSSQLConnectorFunc) Connect(ctx context.Context) (NetConner, error) {
	return fn(ctx)
}

// ReadPreloginFunc defines the type of the func that reads the prelogin packet.
// The production version is implemented by mssql.ReadPreloginWithPacketType.
type ReadPreloginFunc func(
	tdsBuffer interface{},
	pktType interface{}) (map[uint8][]byte, error)

// WritePreloginFunc defines the type of the func that writes the prelogin
// packet. The production version is implemented by
// mssql.WritePreloginWithPacketType.
type WritePreloginFunc func(
	tdsBuffer interface{},
	fields map[uint8][]byte,
	pktType interface{}) error

// ReadLoginFunc defins the type of the func that reads the client's login
// packet.  The production version is implemented by:
//     mssql.ReadLogin(r *TdsBuffer) (*Login, error)
// Note that, in order to avoid a concrete dependency on mssql in this package,
// we must replace TdsBuffer with ReadNextPacketer and *Login with interface{}.
// That interface{} will then be case back to a *Login by the receiving code
// inside the driver package.
type ReadLoginFunc func(r ReadNextPacketer) (interface{}, error)

// TdsBufferCtor represents the constructor of a TdsBuffer, in a form
// suitable for unit tests.
type TdsBufferCtor func(transport io.ReadWriteCloser) ReadNextPacketer

// ReadNextPacketer is an interface that represents the one method on a
// TdsBuffer that we use -- ReadNextPacket().  It allows us to create a mockable
// type to represent a TdsBuffer, and is used together with TdsBufferCtor.
type ReadNextPacketer interface {
	ReadNextPacket() error
}
