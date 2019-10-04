package mock

import (
	"net"

	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// ConnectorMock has a `Connect` method that matches the signature of the
// Connector func type
type ConnectorMock struct {
	mock.Mock
}

// Connect mocks the Connector func type
func (c *ConnectorMock) Connect(
	clientConn net.Conn,
	secrets connector.CredentialValuesByID,
) (backendConn net.Conn, err error) {
	args := c.Called()

	// check for nil because the mock package is unable type assert nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(net.Conn), args.Error(1)
}

// NewConnector creates mock with the `Connect` method that matches the signature
// of the Connector func type
func NewConnector() *ConnectorMock {
	return new(ConnectorMock)
}
