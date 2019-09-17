package mock

import (
	"net"

	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
)

type connectorMock struct {
	mock.Mock
}

func (c *connectorMock) Connect(clientConn net.Conn, secrets plugin.SecretsByID) (backendConn net.Conn, err error) {
	args := c.Called()

	// check for nil because the mock package is unable type assert nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(net.Conn), args.Error(1)
}

// NewConnector creates mock with the `Connect` method that matches the signature
// of the Connector func type
func NewConnector() *connectorMock {
	return new(connectorMock)
}
