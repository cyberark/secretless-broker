package mock

import (
	"context"

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

