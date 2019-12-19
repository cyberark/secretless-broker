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

type ReadPreloginFunc func(
	tdsBuffer interface{},
	pktType interface{}) (map[uint8][]byte, error)
type WritePreloginFunc func(
	tdsBuffer  interface{},
	fields map[uint8][]byte,
	pktType interface{}) error

// TdsBufferCtor represents the constructor of a TdsBuffer, in a form
// suitable for unit tests.
type TdsBufferCtor func(transport io.ReadWriteCloser) ReadNextPacketer

// ReadNextPacketer is an interface that represents the one method on a
// TdsBuffer that we use -- ReadNextPacket().  It allows us to create a mockable
// type to represent a TdsBuffer, and is used together with TdsBufferCtor.
type ReadNextPacketer interface {
	ReadNextPacket() error
}
