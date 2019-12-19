package types

import (
	"context"
	"io"
	"net"

	mssql "github.com/denisenkom/go-mssqldb"
)

// NewMSSQLConnectorFunc represents the constructor of an mssqlConnector. It
// exists so that its production implementation (mssql.NewConnector) can be
// swapped out in unit tests.  Note we keep MSSQL in the name to prevent
// confusion with the Secretless Connector.
type NewMSSQLConnectorFunc func(dsn string) (MSSQLConnector, error)


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

// MSSQLConnectorFunc lets us treat a pure function (that matches the "Connect"
// signature) as an MSSQLConnector interface.
type MSSQLConnectorFunc func(context.Context) (NetConner, error)

// Connect implements the MSSQLConnector interface on a pure function.
func (fn MSSQLConnectorFunc) Connect(ctx context.Context) (NetConner, error) {
	return fn(ctx)
}


// ReadPreloginFunc defines...
type ReadPreloginFunc func(*mssql.TdsBuffer, mssql.PacketType) (map[uint8][]byte, error)
type WritePreloginFunc func(*mssql.TdsBuffer, map[uint8][]byte, mssql.PacketType) error

type NewTdsBufferFunc func(uint16, io.ReadWriteCloser) *mssql.TdsBuffer
