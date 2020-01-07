package mock

import (
	"context"
	"io"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
)

// NewSuccessfulMSSQLConnectorCtor returns an MSSQLConnectorCtor that always
// succeeds.
func NewSuccessfulMSSQLConnectorCtor(
	fn types.MSSQLConnectorFunc,
) types.MSSQLConnectorCtor {
	return func(dsn string) (types.MSSQLConnector, error) {
		return types.MSSQLConnector(fn), nil
	}
}

// NewFailingMSSQLConnectorCtor returns an MSSQLConnectorCtor that always
// returns the specified error.
func NewFailingMSSQLConnectorCtor(err error) types.MSSQLConnectorCtor {
	return func(dsn string) (types.MSSQLConnector, error) {
		return nil, err
	}
}

// NewSuccessfulMSSQLConnector returns an MSSQLConnector double whose Connect
// method always succeeds.
func NewSuccessfulMSSQLConnector(
	fn func(context.Context) (net.Conn, error),
) types.MSSQLConnector {
	return types.MSSQLConnectorFunc(fn)
}

// NewFailingMSSQLConnector returns an MSSQLConnector double whose Connect
// method always fails.
func NewFailingMSSQLConnector(err error) types.MSSQLConnector {
	rawFunc := func(context.Context) (net.Conn, error) {
		return nil, err
	}
	return types.MSSQLConnectorFunc(rawFunc)
}

// SuccessfulReadPrelogin is a double for a ReadPreloginFunc that always
// succeeds.
func SuccessfulReadPrelogin(io.ReadWriteCloser, uint8) (map[uint8][]byte, error) {
	return nil, nil
}

// SuccessfulWritePrelogin is a double for a WritePreloginFunc that always
// succeeds.
func SuccessfulWritePrelogin(io.ReadWriteCloser, map[uint8][]byte, uint8) error {
	return nil
}

// SuccessfulReadLogin is a double for a ReadLoginFunc that always succeeds.
func SuccessfulReadLogin(r io.ReadWriteCloser) (interface{}, error) {
	return struct{}{}, nil
}

// NewNetConn returns a net.Conn double whose behavior we can control.
func NewNetConn(errOnWrite error) *NetConn {
	return &NetConn{errOnWrite: errOnWrite}
}

// NetConn acts as a double of a true network connection, ie, a net.Conn.
// TODO: This will need to be upgraded to have more granularity.  For example,
//   to handle cases where sending the authentication OK message works, but
//   sending an error fails.  Etc.
type NetConn struct {
	net.Conn
	errOnWrite error
}

// Write "writes" bytes to our fake net.Conn.
func (n *NetConn) Write([]byte) (numBytes int, err error) {
	return 1, n.errOnWrite
}

// FakeTdsBufferCtor returns the ReadWriteCloser passed in.
func FakeTdsBufferCtor(r io.ReadWriteCloser) io.ReadWriteCloser {
	return r
}
