package mock

import (
	"context"
	"net"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
)

func NewSuccessfulMSSQLConnectorConstructor(
	fn types.MSSQLConnectorFunc,
) types.NewMSSQLConnectorFunc{
	return func(dsn string) (types.MSSQLConnector, error) {
		return types.MSSQLConnector(fn), nil
	}
}

func NewFailingMSSQLConnectorConstructor(err error) types.NewMSSQLConnectorFunc{
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

func SuccessfulReadPrelogin(*mssql.TdsBuffer, mssql.PacketType) (map[uint8][]byte, error) {
	return nil, nil
}

func SuccessfulWritePrelogin(*mssql.TdsBuffer, map[uint8][]byte, mssql.PacketType) error {
	return nil
}

func NewNetConn() *NetConn {
	return &NetConn{}
}

type NetConn struct {
	net.Conn
}

func (n *NetConn) NetConn() net.Conn {
	return n
}
