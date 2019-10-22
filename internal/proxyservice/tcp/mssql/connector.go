package mssql

import (
	"context"
	"fmt"
	"net"

	mssql "github.com/denisenkom/go-mssqldb"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

type SingleUseConnector struct {
	logger log.Logger
}

func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	// TODO: Magic here

	// using DSN because it seems gomssql forces us to
	c, err := mssql.NewConnector(
		fmt.Sprintf(
			"sqlserver://%s:%s@%s",
			credentialValuesByID["user"],
			credentialValuesByID["password"],
			credentialValuesByID["address"],
			),
		)
	if err != nil {
		connector.logger.Errorf("bad connector: %s", err)
		return nil, err
	}

	_conn, err := c.Connect(context.Background())
	if err != nil {
		connector.logger.Errorf("bad connect: %s", err)
		return nil, err
	}
	conn := _conn.(*mssql.Conn)
	backEndConn := conn.NetConn()

	connector.logger.Infof("backend connection ready")

	err = mssql.PerformFakeHandshakeForClient(clientConn)
	if err != nil {
		connector.logger.Errorf("bad PerformFakeHandshakeForClient: %s", err)
		return nil, err
	}

	return backEndConn, nil
}
