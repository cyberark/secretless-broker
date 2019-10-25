package template

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

type SingleUseConnector struct {
	logger log.Logger
}

// This function receives a connection to the client, and opens a connection to the target using the client's connection
// and the credentials provided in credentialValuesByID
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	// TODO: add logic according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#tcp-connector
	// tcp/pg/connector.go is a good example.

	var err error
	return nil, err
}
