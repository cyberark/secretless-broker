package pg

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthConn struct {
	writeData []byte
	readData  []byte
	readPos   int
	closed    bool
}

func newMockAuthConn() *mockAuthConn {
	return &mockAuthConn{
		writeData: []byte{},
		readData:  []byte{},
		readPos:   0,
	}
}

func (m *mockAuthConn) Read(b []byte) (int, error) {
	if m.readPos >= len(m.readData) {
		return 0, fmt.Errorf("no more data to read")
	}
	n := copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockAuthConn) Write(b []byte) (int, error) {
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockAuthConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockAuthConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5432}
}

func (m *mockAuthConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}
}

func (m *mockAuthConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockAuthConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockAuthConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockBackendConn struct {
	*mockAuthConn
}

func newMockBackendConn() *mockBackendConn {
	return &mockBackendConn{
		mockAuthConn: newMockAuthConn(),
	}
}

func TestHandleSSL_NoTLS(t *testing.T) {
	connector := &SingleUseConnector{
		connectionDetails: &ConnectionDetails{
			SSLOptions: map[string]string{
				"sslmode": "disable",
			},
		},
	}

	err := connector.handleSSL()
	assert.NoError(t, err)
}

func TestHandleSSL_SSLAllowed(t *testing.T) {
	mockConn := newMockBackendConn()
	mockConn.readData = []byte{'S'}

	connector := &SingleUseConnector{
		backendConn: mockConn,
		connectionDetails: &ConnectionDetails{
			SSLOptions: map[string]string{
				"sslmode": "require",
			},
		},
	}

	err := connector.handleSSL()
	assert.Error(t, err)

	assert.True(t, len(mockConn.writeData) > 0)
	msg := protocol.NewMessageBuffer(mockConn.writeData)
	length, readErr := msg.ReadInt32()
	require.NoError(t, readErr)
	assert.Equal(t, int32(8), length)
	code, readErr := msg.ReadInt32()
	require.NoError(t, readErr)
	assert.Equal(t, protocol.SSLRequestCode, code)
}

func TestHandleSSL_SSLNotAllowed(t *testing.T) {
	mockConn := newMockBackendConn()
	mockConn.readData = []byte{'N'}

	connector := &SingleUseConnector{
		backendConn: mockConn,
		connectionDetails: &ConnectionDetails{
			SSLOptions: map[string]string{
				"sslmode": "require",
			},
		},
	}

	err := connector.handleSSL()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not allow SSL connections")
	assert.True(t, mockConn.closed)
}

func TestHandleSSL_ReadError(t *testing.T) {
	mockConn := newMockBackendConn()

	connector := &SingleUseConnector{
		backendConn: mockConn,
		connectionDetails: &ConnectionDetails{
			SSLOptions: map[string]string{
				"sslmode": "require",
			},
		},
	}

	err := connector.handleSSL()
	assert.Error(t, err)
}

func TestConnectionDetails_Address(t *testing.T) {
	tests := []struct {
		name     string
		details  *ConnectionDetails
		expected string
	}{
		{
			name: "custom host and port",
			details: &ConnectionDetails{
				Host: "example.com",
				Port: "5433",
			},
			expected: "example.com:5433",
		},
		{
			name: "default port",
			details: &ConnectionDetails{
				Host: "localhost",
				Port: DefaultPostgresPort,
			},
			expected: "localhost:5432",
		},
		{
			name: "IPv6 address",
			details: &ConnectionDetails{
				Host: "::1",
				Port: "5432",
			},
			expected: "[::1]:5432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.details.Address())
		})
	}
}

func TestConnectionDetails_OptionsHandling(t *testing.T) {
	connector := &SingleUseConnector{
		connectionDetails: &ConnectionDetails{
			Host:     "localhost",
			Port:     "5432",
			Username: "testuser",
			Password: "testpass",
			Options: map[string]string{
				"application_name": "backend_app",
				"client_encoding":  "LATIN1",
			},
			SSLOptions: map[string]string{
				"sslmode": "disable",
			},
		},
		options: map[string]string{
			"client_encoding":  "UTF8",
			"application_name": "client_app",
		},
		databaseName: "testdb",
	}

	options := map[string]string{}
	for k, v := range connector.connectionDetails.Options {
		options[k] = v
	}
	for k, v := range connector.options {
		options[k] = v
	}

	assert.Equal(t, "client_app", options["application_name"])
	assert.Equal(t, "UTF8", options["client_encoding"])
}

func TestConnectionDetails_EmptyCredentials(t *testing.T) {
	tests := []struct {
		name        string
		details     *ConnectionDetails
		shouldError bool
	}{
		{
			name: "empty username",
			details: &ConnectionDetails{
				Host:     "localhost",
				Port:     "5432",
				Username: "",
				Password: "testpass",
			},
			shouldError: true,
		},
		{
			name: "empty password",
			details: &ConnectionDetails{
				Host:     "localhost",
				Port:     "5432",
				Username: "testuser",
				Password: "",
			},
			shouldError: true,
		},
		{
			name: "valid credentials",
			details: &ConnectionDetails{
				Host:     "localhost",
				Port:     "5432",
				Username: "testuser",
				Password: "testpass",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldError {
				assert.True(t, tt.details.Username == "" || tt.details.Password == "")
			} else {
				assert.NotEmpty(t, tt.details.Username)
				assert.NotEmpty(t, tt.details.Password)
			}
		})
	}
}

func TestConnectionDetails_SSLOptions(t *testing.T) {
	details := &ConnectionDetails{
		SSLOptions: map[string]string{
			"sslmode":     "require",
			"sslrootcert": "/path/to/cert",
			"sslkey":      "/path/to/key",
		},
	}

	assert.Equal(t, "require", details.SSLOptions["sslmode"])
	assert.Equal(t, "/path/to/cert", details.SSLOptions["sslrootcert"])
	assert.Equal(t, "/path/to/key", details.SSLOptions["sslkey"])
	assert.Len(t, details.SSLOptions, 3)
}
