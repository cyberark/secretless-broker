package mysql

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql/protocol"
	"github.com/stretchr/testify/assert"
)

type mockNetConn struct {
	readData    []byte
	writeBuffer []byte
	readErr     error
	writeErr    error
	readPos     int
	closed      bool
}

func (m *mockNetConn) Read(b []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}
	n = copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockNetConn) Write(b []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.writeBuffer = append(m.writeBuffer, b...)
	return len(b), nil
}

func (m *mockNetConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3306}
}
func (m *mockNetConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 33060}
}
func (m *mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error { return nil }

func createValidHandshakePacket() []byte {
	packet := make([]byte, 4+1+19+8+1+2+1+2+10+12+1)

	// Header
	packet[0] = byte(len(packet) - 4)
	packet[1] = 0
	packet[2] = 0
	packet[3] = 0 // sequence id

	// Payload
	pos := 4
	packet[pos] = 10 // protocol version
	pos++

	// Server version (null terminated)
	copy(packet[pos:], []byte("5.7.0-test\x00"))
	pos += 11

	// Connection ID (4 bytes)
	pos += 4

	// Auth plugin data part 1 (8 bytes)
	pos += 8

	// Filler (1 byte)
	pos++

	// Capability flags lower 2 bytes - FIX: proper uint32 to byte conversion
	clientSSL := uint32(protocol.ClientSSL)
	packet[pos] = byte(clientSSL & 0xFF)
	packet[pos+1] = byte((clientSSL >> 8) & 0xFF)
	pos += 2

	// Character set (1 byte)
	pos++

	// Status flags (2 bytes)
	pos += 2

	// Capability flags upper 2 bytes
	packet[pos] = byte((clientSSL >> 16) & 0xFF)
	packet[pos+1] = byte((clientSSL >> 24) & 0xFF)
	pos += 2

	// Auth plugin data length (1 byte)
	packet[pos] = 21
	pos++

	// Reserved (10 bytes)
	pos += 10

	// Auth plugin data part 2 (12 bytes)
	pos += 12

	// Auth plugin name (null terminated)
	packet[pos] = 0

	return packet
}

func TestNewAuthenticationHandshake(t *testing.T) {
	clientConn := NewClientConnection(&mockNetConn{})
	backendConn := NewBackendConnection(&mockNetConn{})
	connDetails := &ConnectionDetails{
		Username: "testuser",
		Password: "testpass",
	}

	handshake := NewAuthenticationHandshake(clientConn, backendConn, connDetails)

	assert.Equal(t, clientConn, handshake.clientConn)
	assert.Equal(t, backendConn, handshake.backendConn)
	assert.Equal(t, connDetails, handshake.connectionDetails)
	assert.Nil(t, handshake.err)
}

func TestAuthenticationHandshake_AuthenticatedBackendConn(t *testing.T) {
	mockBackend := &mockNetConn{}
	backendConn := NewBackendConnection(mockBackend)
	handshake := AuthenticationHandshake{
		backendConn: backendConn,
	}

	result := handshake.AuthenticatedBackendConn()
	assert.Equal(t, mockBackend, result)
}

func TestAuthenticationHandshake_clientRequestedSSL(t *testing.T) {
	tests := []struct {
		name           string
		sslOptions     map[string]string
		expectedResult bool
	}{
		{
			name:           "SSL enabled",
			sslOptions:     map[string]string{"sslmode": "require"},
			expectedResult: true,
		},
		{
			name:           "SSL disabled",
			sslOptions:     map[string]string{"sslmode": "disable"},
			expectedResult: false,
		},
		{
			name:           "No SSL options",
			sslOptions:     map[string]string{"sslmode": "disable"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handshake := AuthenticationHandshake{
				connectionDetails: &ConnectionDetails{
					SSLOptions: tt.sslOptions,
				},
			}

			result := handshake.clientRequestedSSL()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func createValidHandshakeResponse() []byte {
	// Prawidłowy HandshakeResponse41 packet
	payload := make([]byte, 0, 100)

	// Capability flags (4 bytes) - CLIENT_PROTOCOL_41 | CLIENT_SECURE_CONNECTION
	cap := protocol.ClientProtocol41 | protocol.ClientSecureConnection
	payload = append(payload, byte(cap), byte(cap>>8), byte(cap>>16), byte(cap>>24))

	// Max packet size (4 bytes)
	payload = append(payload, 0xff, 0xff, 0xff, 0x00)

	// Character set (1 byte)
	payload = append(payload, 33)

	// Reserved (23 bytes)
	payload = append(payload, make([]byte, 23)...)

	// Username null terminated
	payload = append(payload, []byte("testuser\x00")...)

	// Auth response length (1 byte) + auth response (20 bytes for mysql_native_password)
	payload = append(payload, 20)
	payload = append(payload, make([]byte, 20)...) // dummy auth response

	// Database name null terminated (optional)
	payload = append(payload, []byte("testdb\x00")...)

	// Auth plugin name null terminated
	payload = append(payload, []byte("mysql_native_password\x00")...)

	// Create packet with header
	packet := make([]byte, 4+len(payload))

	// Header: length (3 bytes) + sequence ID (1 byte)
	packet[0] = byte(len(payload))
	packet[1] = byte(len(payload) >> 8)
	packet[2] = byte(len(payload) >> 16)
	packet[3] = 1 // sequence ID

	// Payload
	copy(packet[4:], payload)

	return packet
}

func TestAuthenticationHandshake_serverSupportsSSL(t *testing.T) {
	tests := []struct {
		name                string
		serverCapabilities  uint32
		expectedSupportsSSL bool
	}{
		{
			name:                "Server supports SSL",
			serverCapabilities:  protocol.ClientSSL,
			expectedSupportsSSL: true,
		},
		{
			name:                "Server doesn't support SSL",
			serverCapabilities:  protocol.ClientProtocol41,
			expectedSupportsSSL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handshake := AuthenticationHandshake{
				serverHandshake: &protocol.HandshakeV10{
					ServerCapabilities: tt.serverCapabilities,
				},
			}

			result := handshake.serverSupportsSSL()
			assert.Equal(t, tt.expectedSupportsSSL, result)
		})
	}
}

func TestAuthenticationHandshake_readServerHandshake_Success(t *testing.T) {
	mockBackend := &mockNetConn{
		readData: createValidHandshakePacket(),
	}
	backendConn := NewBackendConnection(mockBackend)

	handshake := AuthenticationHandshake{
		backendConn: backendConn,
	}

	handshake.readServerHandshake()

	assert.NoError(t, handshake.err)
	assert.NotNil(t, handshake.rawServerHandshake)
	assert.NotNil(t, handshake.serverHandshake)
}

func TestAuthenticationHandshake_readServerHandshake_ReadError(t *testing.T) {
	mockBackend := &mockNetConn{
		readErr: errors.New("read error"),
	}
	backendConn := NewBackendConnection(mockBackend)

	handshake := AuthenticationHandshake{
		backendConn: backendConn,
	}

	handshake.readServerHandshake()

	assert.Error(t, handshake.err)
	assert.Contains(t, handshake.err.Error(), "read error")
}

func TestAuthenticationHandshake_readClientHandshakeResponse_Success(t *testing.T) {
	mockClient := &mockNetConn{
		readData: createValidHandshakeResponse(),
	}
	clientConn := NewClientConnection(mockClient)

	handshake := AuthenticationHandshake{
		clientConn: clientConn,
	}

	handshake.readClientHandshakeResponse()

	assert.NoError(t, handshake.err)
	assert.NotNil(t, handshake.clientHandshakeResponse)
}

func TestAuthenticationHandshake_overrideClientCapabilities(t *testing.T) {
	handshake := AuthenticationHandshake{
		clientHandshakeResponse: &protocol.HandshakeResponse41{
			CapabilityFlags: protocol.ClientProtocol41,
		},
		connectionDetails: &ConnectionDetails{
			SSLOptions: map[string]string{"enable": "false"},
		},
	}

	originalFlags := handshake.clientHandshakeResponse.CapabilityFlags

	handshake.overrideClientCapabilities()

	assert.NoError(t, handshake.err)
	assert.True(t, (handshake.clientHandshakeResponse.CapabilityFlags&protocol.ClientPluginAuth) != 0)
	assert.True(t, (handshake.clientHandshakeResponse.CapabilityFlags&protocol.ClientSecureConnection) != 0)
	assert.NotEqual(t, originalFlags, handshake.clientHandshakeResponse.CapabilityFlags)
}

func TestAuthenticationHandshake_ErrorGuards(t *testing.T) {
	handshake := AuthenticationHandshake{
		err: errors.New("existing error"),
	}

	// All these methods should return early due to existing error
	handshake.readServerHandshake()
	handshake.writeHandshakeToClient()
	handshake.validateServerSSL()
	handshake.readClientHandshakeResponse()
	handshake.overrideClientCapabilities()
	handshake.injectCredentials()
	handshake.handleClientSSLRequest()
	handshake.writeClientHandshakeResponseToBackend()
	handshake.verifyAndProxyOkResponse()
	handshake.handleBackendAuthResponse()

	// Verify the original error is preserved
	assert.Equal(t, "existing error", handshake.err.Error())
}

func TestAuthenticationHandshake_readPacket_Success(t *testing.T) {
	testPacket := []byte{0x01, 0x00, 0x00, 0x00, 0x01}
	mockConn := &mockNetConn{readData: testPacket}
	conn := NewClientConnection(mockConn)

	handshake := AuthenticationHandshake{}

	result := handshake.readPacket(conn)

	assert.NotNil(t, result)
	assert.Equal(t, testPacket, []byte(result))
}

func TestAuthenticationHandshake_readPacket_Error(t *testing.T) {
	mockConn := &mockNetConn{readErr: errors.New("read error")}
	conn := NewClientConnection(mockConn)

	handshake := AuthenticationHandshake{}

	result := handshake.readPacket(conn)

	assert.Error(t, handshake.err)
	assert.Nil(t, result)
}
