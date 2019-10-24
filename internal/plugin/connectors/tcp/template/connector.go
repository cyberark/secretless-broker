package template

import (
	"net"

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
	// TODO: add logic according to
	//  https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#tcp-connector

	return nil, nil
}
