package mssql

import (
	"context"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"

	mssql "github.com/denisenkom/go-mssqldb"
)

type SingleUseConnector struct {
	logger log.Logger
}

func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {

	// Using DSN (Data Source Name) string because gomssql forces us to.
	//
	// NOTE: Secretless has some unfortunate naming collisions with the
	// go-mssqldb driver package.  The driver package has its own concept of a
	// "connector", and its connectors also have a "Connect" method.
	driverConnector, err := mssql.NewConnector(
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

	driverConn, err := driverConnector.Connect(context.Background())
	if err != nil {
		connector.logger.Errorf("bad connect: %s", err)
		return nil, err
	}

	// Verify the driverConn is an mssql driverConn object and get its underlying transport
	mssqlConn := driverConn.(*mssql.Conn)
	backEndConn := mssqlConn.NetConn()

	connector.logger.Infof("backend driverConn ready")

	// TODO: add a comment on this - why do we need it?
	err = PerformFakeHandshakeForClient(clientConn)
	if err != nil {
		connector.logger.Errorf("bad PerformFakeHandshakeForClient: %s", err)
		return nil, err
	}

	return backEndConn, nil
}
