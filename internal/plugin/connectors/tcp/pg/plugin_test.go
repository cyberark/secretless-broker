package pg

import (
	"net"
	"testing"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResources struct {
	logger log.Logger
}

func (m *mockResources) Logger() log.Logger {
	return m.logger
}

func (m *mockResources) Config() []byte {
	return []byte{}
}

func TestPluginInfo(t *testing.T) {
	info := PluginInfo()

	require.NotNil(t, info)
	assert.Equal(t, "0.1.0", info["pluginAPIVersion"])
	assert.Equal(t, "connector.tcp", info["type"])
	assert.Equal(t, "pg", info["id"])
	assert.Equal(t, "returns an authenticated connection to a PostgreSQL database", info["description"])
}

func TestGetTCPPlugin(t *testing.T) {
	plugin := GetTCPPlugin()

	assert.NotNil(t, plugin)
}

func TestNewConnector(t *testing.T) {
	resources := &mockResources{
		logger: &mockLogger{},
	}

	conn := NewConnector(resources)

	assert.NotNil(t, conn)
}

func TestNewConnector_Connect(t *testing.T) {
	resources := &mockResources{
		logger: &mockLogger{},
	}

	tcpConnector := NewConnector(resources)

	mockClientConn := &mockConnForConnect{
		mockAuthConn: newMockAuthConn(),
	}

	credentials := connector.CredentialValuesByID{
		"host":     []byte("localhost"),
		"port":     []byte("5432"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	backendConn, err := tcpConnector.Connect(mockClientConn, credentials)

	assert.Error(t, err)
	assert.Nil(t, backendConn)
}

func TestNewConnector_Integration(t *testing.T) {
	tests := []struct {
		name        string
		resources   connector.Resources
		clientConn  net.Conn
		credentials connector.CredentialValuesByID
		expectError bool
		expectPanic bool
	}{
		{
			name: "valid resources and credentials",
			resources: &mockResources{
				logger: &mockLogger{},
			},
			clientConn: &mockConnForConnect{
				mockAuthConn: newMockAuthConn(),
			},
			credentials: connector.CredentialValuesByID{
				"host":     []byte("localhost"),
				"port":     []byte("5432"),
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
			expectError: true,
			expectPanic: false,
		},
		{
			name: "nil logger",
			resources: &mockResources{
				logger: nil,
			},
			clientConn: &mockConnForConnect{
				mockAuthConn: newMockAuthConn(),
			},
			credentials: connector.CredentialValuesByID{
				"host":     []byte("localhost"),
				"port":     []byte("5432"),
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
			expectError: false,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					tcpConnector := NewConnector(tt.resources)
					tcpConnector.Connect(tt.clientConn, tt.credentials)
				})
				return
			}

			tcpConnector := NewConnector(tt.resources)
			backendConn, err := tcpConnector.Connect(tt.clientConn, tt.credentials)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, backendConn)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backendConn)
			}
		})
	}
}

func TestNewConnector_LoggerInitialization(t *testing.T) {
	resources := &mockResources{
		logger: &mockLogger{},
	}

	tcpConnector := NewConnector(resources)

	mockClientConn := &mockConnForConnect{
		mockAuthConn: newMockAuthConn(),
	}

	credentials := connector.CredentialValuesByID{
		"host":     []byte("localhost"),
		"port":     []byte("5432"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	_, err := tcpConnector.Connect(mockClientConn, credentials)

	assert.Error(t, err)
}

func TestPluginInfo_AllFields(t *testing.T) {
	info := PluginInfo()

	expectedFields := []string{"pluginAPIVersion", "type", "id", "description"}
	for _, field := range expectedFields {
		assert.Contains(t, info, field)
		assert.NotEmpty(t, info[field])
	}

	assert.Len(t, info, 4)
}
