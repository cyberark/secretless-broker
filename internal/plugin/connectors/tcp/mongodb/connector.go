package mongodb

import (
	"context"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// SingleUseConnector is passed the client's net.Conn and the current CredentialValuesById,
// and returns an authenticated net.Conn to the target service
type SingleUseConnector struct {
	logger log.Logger
}

// Connect receives a connection to the client, and opens a connection to the target using the client's connection
// and the credentials provided in credentialValuesByID
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	connDetails, _ := NewConnectionDetails(credentialValuesByID)

	host := net.JoinHostPort(connDetails.Host, fmt.Sprintf("%d", connDetails.Port))
	dialer := newProxyDialer()
	backendConn, err := dialer.DialContext(context.Background(), "tcp", host)
	if err != nil {
		return nil, err
	}

	return backendConn, nil
}
