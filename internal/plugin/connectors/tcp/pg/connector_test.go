package pg

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                       {}
func (m *mockLogger) Debugf(format string, args ...interface{})       {}
func (m *mockLogger) Debugln(args ...interface{})                     {}
func (m *mockLogger) Info(args ...interface{})                        {}
func (m *mockLogger) Infof(format string, args ...interface{})        {}
func (m *mockLogger) Infoln(args ...interface{})                      {}
func (m *mockLogger) Warn(args ...interface{})                        {}
func (m *mockLogger) Warnf(format string, args ...interface{})        {}
func (m *mockLogger) Warnln(args ...interface{})                      {}
func (m *mockLogger) Error(args ...interface{})                       {}
func (m *mockLogger) Errorf(format string, args ...interface{})       {}
func (m *mockLogger) Errorln(args ...interface{})                     {}
func (m *mockLogger) Panic(args ...interface{})                       {}
func (m *mockLogger) Panicf(format string, args ...interface{})       {}
func (m *mockLogger) Panicln(args ...interface{})                     {}
func (m *mockLogger) Fatal(args ...interface{})                       {}
func (m *mockLogger) Fatalf(format string, args ...interface{})       {}
func (m *mockLogger) Fatalln(args ...interface{})                     {}
func (m *mockLogger) CopyWith(prefix string, isDebug bool) log.Logger { return m }
func (m *mockLogger) DebugEnabled() bool                              { return false }
func (m *mockLogger) Prefix() string                                  { return "" }

func TestSingleUseConnector_Abort_ProtocolError(t *testing.T) {
	mockConn := newMockAuthConn()

	suc := &SingleUseConnector{
		clientConn: mockConn,
	}

	pgErr := &protocol.Error{
		Severity: protocol.ErrorSeverityFatal,
		Code:     "28P01",
		Message:  "password authentication failed",
	}

	suc.abort(pgErr)

	assert.True(t, len(mockConn.writeData) > 0)
	assert.Equal(t, protocol.ErrorMessageType, mockConn.writeData[0])
}

func TestSingleUseConnector_Abort_GenericError(t *testing.T) {
	mockConn := newMockAuthConn()

	suc := &SingleUseConnector{
		clientConn: mockConn,
	}

	err := errors.New("generic error message")
	suc.abort(err)

	assert.True(t, len(mockConn.writeData) > 0)
	msg := protocol.NewMessageBuffer(mockConn.writeData[5:])

	_, _ = msg.ReadByte()
	severity, _ := msg.ReadString()
	assert.Equal(t, protocol.ErrorSeverityFatal, severity)

	_, _ = msg.ReadByte()
	code, _ := msg.ReadString()
	assert.Equal(t, protocol.ErrorCodeInternalError, code)

	_, _ = msg.ReadByte()
	message, _ := msg.ReadString()
	assert.Equal(t, "generic error message", message)
}

func TestSingleUseConnector_Abort_NilConnection(t *testing.T) {
	suc := &SingleUseConnector{
		clientConn: nil,
	}

	err := errors.New("test error")

	assert.NotPanics(t, func() {
		suc.abort(err)
	})
}

func TestSingleUseConnector_SSLNotSupported(t *testing.T) {
	mockConn := newMockAuthConn()

	suc := &SingleUseConnector{
		clientConn: mockConn,
	}

	suc.sslNotSupported()

	assert.Equal(t, 1, len(mockConn.writeData))
	assert.Equal(t, protocol.SSLNotAllowed, mockConn.writeData[0])
}

func TestSingleUseConnector_SSLNotSupported_NilConnection(t *testing.T) {
	suc := &SingleUseConnector{
		clientConn: nil,
	}

	assert.NotPanics(t, func() {
		suc.sslNotSupported()
	})
}

func TestSingleUseConnector_Abort_WriteError(t *testing.T) {
	mockConn := &mockAuthConn{
		writeData: []byte{},
		readData:  []byte{},
		closed:    true,
	}

	suc := &SingleUseConnector{
		clientConn: mockConn,
	}

	err := errors.New("test error")

	assert.NotPanics(t, func() {
		suc.abort(err)
	})
}

type mockConnWithDelay struct {
	*mockAuthConn
	readDelay time.Duration
}

func (m *mockConnWithDelay) Read(b []byte) (int, error) {
	time.Sleep(m.readDelay)
	return m.mockAuthConn.Read(b)
}

type mockConnForConnect struct {
	*mockAuthConn
	startupReturnsEOF      bool
	startupReturnsError    bool
	connectBackendFails    bool
	parseStartupReturnsSSL bool
	secondSSLRequest       bool
	missingDatabase        bool
}

func (m *mockConnForConnect) Read(b []byte) (int, error) {
	if m.startupReturnsEOF {
		return 0, io.EOF
	}
	if m.startupReturnsError {
		return 0, errors.New("startup error")
	}
	if m.parseStartupReturnsSSL {
		if m.secondSSLRequest {
			msg := protocol.NewMessageBuffer([]byte{})
			msg.WriteInt32(8)
			msg.WriteInt32(protocol.SSLRequestCode)
			copy(b, msg.Bytes())
			return len(msg.Bytes()), nil
		}
		m.secondSSLRequest = true
		msg := protocol.NewMessageBuffer([]byte{})
		msg.WriteInt32(8)
		msg.WriteInt32(protocol.SSLRequestCode)
		copy(b, msg.Bytes())
		return len(msg.Bytes()), nil
	}

	options := map[string]string{}
	if !m.missingDatabase {
		options["database"] = "testdb"
	}
	startupMsg := protocol.CreateStartupMessage(
		protocol.ProtocolVersion,
		"testuser",
		"testdb",
		options,
	)
	copy(b, startupMsg)
	return len(startupMsg), nil
}

func TestSingleUseConnector_Connect_StartupEOF(t *testing.T) {
	mockConn := &mockConnForConnect{
		mockAuthConn:      newMockAuthConn(),
		startupReturnsEOF: true,
	}

	suc := &SingleUseConnector{
		logger: &mockLogger{},
	}

	credentials := connector.CredentialValuesByID{
		"host":     []byte("localhost"),
		"port":     []byte("5432"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	conn, err := suc.Connect(mockConn, credentials)

	assert.Equal(t, io.EOF, err)
	assert.Nil(t, conn)
	assert.Equal(t, 0, len(mockConn.writeData))
}

func TestSingleUseConnector_Connect_StartupError(t *testing.T) {
	mockConn := &mockConnForConnect{
		mockAuthConn:        newMockAuthConn(),
		startupReturnsError: true,
	}

	suc := &SingleUseConnector{
		logger: &mockLogger{},
	}

	credentials := connector.CredentialValuesByID{
		"host":     []byte("localhost"),
		"port":     []byte("5432"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	conn, err := suc.Connect(mockConn, credentials)

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.True(t, len(mockConn.writeData) > 0)
}

func TestSingleUseConnector_Connect_InvalidAddress(t *testing.T) {
	mockConn := &mockConnForConnect{
		mockAuthConn: newMockAuthConn(),
	}

	suc := &SingleUseConnector{
		logger: &mockLogger{},
	}

	credentials := connector.CredentialValuesByID{
		"address":  []byte("invalid:address:format:test"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	conn, err := suc.Connect(mockConn, credentials)

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.True(t, len(mockConn.writeData) > 0)
}

func TestSingleUseConnector_Connect_MissingDatabase(t *testing.T) {
	mockConn := &mockConnForConnect{
		mockAuthConn:    newMockAuthConn(),
		missingDatabase: true,
	}

	suc := &SingleUseConnector{
		logger: &mockLogger{},
	}

	credentials := connector.CredentialValuesByID{
		"host":     []byte("localhost"),
		"port":     []byte("5432"),
		"username": []byte("testuser"),
		"password": []byte("testpass"),
	}

	conn, err := suc.Connect(mockConn, credentials)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no 'database' found")
	assert.Nil(t, conn)
}

func TestSingleUseConnector_Connect_AllPaths(t *testing.T) {
	tests := []struct {
		name        string
		mockConn    net.Conn
		credentials connector.CredentialValuesByID
		expectError bool
		checkAbort  bool
	}{
		{
			name: "successful startup",
			mockConn: &mockConnForConnect{
				mockAuthConn: newMockAuthConn(),
			},
			credentials: connector.CredentialValuesByID{
				"host":     []byte("localhost"),
				"port":     []byte("5432"),
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
			expectError: true,
			checkAbort:  false,
		},
		{
			name: "EOF during startup",
			mockConn: &mockConnForConnect{
				mockAuthConn:      newMockAuthConn(),
				startupReturnsEOF: true,
			},
			credentials: connector.CredentialValuesByID{
				"host":     []byte("localhost"),
				"port":     []byte("5432"),
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
			expectError: true,
			checkAbort:  false,
		},
		{
			name: "error during startup",
			mockConn: &mockConnForConnect{
				mockAuthConn:        newMockAuthConn(),
				startupReturnsError: true,
			},
			credentials: connector.CredentialValuesByID{
				"host":     []byte("localhost"),
				"port":     []byte("5432"),
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
			expectError: true,
			checkAbort:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suc := &SingleUseConnector{
				logger: &mockLogger{},
			}
			conn, err := suc.Connect(tt.mockConn, tt.credentials)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, conn)

				if tt.checkAbort {
					mockConn := tt.mockConn.(*mockConnForConnect)
					assert.True(t, len(mockConn.writeData) > 0)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
			}
		})
	}
}
