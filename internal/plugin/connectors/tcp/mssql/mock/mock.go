package mock

import (
	"context"
	"io"
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
)

func NewSuccessfulMSSQLConnectorCtor(
	fn types.MSSQLConnectorFunc,
) types.NewMSSQLConnectorFunc{
	return func(dsn string) (types.MSSQLConnector, error) {
		return types.MSSQLConnector(fn), nil
	}
}

func NewFailingMSSQLConnectorCtor(err error) types.NewMSSQLConnectorFunc{
	return func(dsn string) (types.MSSQLConnector, error) {
		return nil, err
	}
}

func NewSuccessfulMSSQLConnector(
	fn func(context.Context) (types.NetConner, error),
) types.MSSQLConnector {
	return types.MSSQLConnectorFunc(fn)
}

func NewFailingMSSQLConnector(err error) types.MSSQLConnector {
	rawFunc := func(context.Context) (types.NetConner, error) {
		return nil, err
	}
	return types.MSSQLConnectorFunc(rawFunc)
}

func SuccessfulReadPrelogin(interface{}, interface{}) (map[uint8][]byte, error) {
	return nil, nil
}

func SuccessfulWritePrelogin(interface{}, map[uint8][]byte, interface{}) error {
	return nil
}

func SuccessfulTdsBufferCtor() types.NewTdsBufferFunc {
	return func(transport io.ReadWriteCloser) types.ReadNextPacketer {
		return NewTdsBuffer(nil)
	}
}

func NewFailingTdsBufferCtor(err error) types.NewTdsBufferFunc {
	return func(transport io.ReadWriteCloser) types.ReadNextPacketer {
		return NewTdsBuffer(err)
	}
}

func NewTdsBuffer(err error) types.ReadNextPacketer {
	return &TdsBuffer{err: err}
}
type TdsBuffer struct{
	err error
}
func (tb *TdsBuffer) ReadNextPacket() error {
	return tb.err
}

func NewNetConn(errOnWrite error) *NetConn {
	return &NetConn{errOnWrite: errOnWrite }
}
// TODO: This will need to be upgraded to have more granularity.  For example,
//   to handle cases where sending the authentication OK message works, but
//   sending an error fails.  Etc.
type NetConn struct {
	net.Conn
	errOnWrite error
}
func (n *NetConn) Write([]byte) (numBytes int, err error) {
	return 1, n.errOnWrite
}
func (n *NetConn) NetConn() net.Conn {
	return n
}
