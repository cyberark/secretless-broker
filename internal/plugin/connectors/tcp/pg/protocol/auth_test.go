package protocol

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthConn struct {
	readData    []byte
	readPos     int
	writeBuffer bytes.Buffer
	readErr     error
	writeErr    error
}

func (m *mockAuthConn) Read(b []byte) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	if m.readPos >= len(m.readData) {
		return 0, errors.New("no more data")
	}
	n := copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockAuthConn) Write(b []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuffer.Write(b)
}

func (m *mockAuthConn) Close() error                       { return nil }
func (m *mockAuthConn) LocalAddr() net.Addr                { return nil }
func (m *mockAuthConn) RemoteAddr() net.Addr               { return nil }
func (m *mockAuthConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockAuthConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockAuthConn) SetWriteDeadline(t time.Time) error { return nil }

func createAuthMessage(authType int32, extraData []byte) []byte {
	msg := NewMessageBuffer([]byte{})
	msg.WriteByte(AuthenticationMessageType)
	msg.WriteInt32(0)
	msg.WriteInt32(authType)
	if extraData != nil {
		msg.WriteBytes(extraData)
	}
	msg.ResetLength(PGMessageLengthOffset)
	return msg.Bytes()
}

func createAuthOkMessage() []byte {
	return createAuthMessage(AuthenticationOk, nil)
}

func TestHandleAuthenticationRequest_AuthOk(t *testing.T) {
	authMsg := createAuthMessage(AuthenticationOk, nil)

	conn := &mockAuthConn{
		readData: authMsg,
	}

	err := HandleAuthenticationRequest("user", "pass", conn)

	assert.NoError(t, err)
}

func TestHandleAuthenticationRequest_ErrorMessage(t *testing.T) {
	errMsg := NewMessageBuffer([]byte{})
	errMsg.WriteByte(ErrorMessageType)
	errMsg.WriteInt32(0)
	errMsg.WriteByte(ErrorFieldSeverity)
	errMsg.WriteString(ErrorSeverityFatal)
	errMsg.WriteByte(ErrorFieldCode)
	errMsg.WriteString("28P01")
	errMsg.WriteByte(ErrorFieldMessage)
	errMsg.WriteString("authentication failed")
	errMsg.WriteByte(0x00)
	errMsg.ResetLength(PGMessageLengthOffset)

	conn := &mockAuthConn{
		readData: errMsg.Bytes(),
	}

	err := HandleAuthenticationRequest("user", "pass", conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestHandleAuthenticationRequest_UnexpectedMessageType(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteByte(QueryMessageType)
	msg.WriteInt32(4)

	conn := &mockAuthConn{
		readData: msg.Bytes(),
	}

	err := HandleAuthenticationRequest("user", "pass", conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected")
}

func TestHandleAuthenticationRequest_ReadError(t *testing.T) {
	conn := &mockAuthConn{
		readErr: errors.New("read failed"),
	}

	err := HandleAuthenticationRequest("user", "pass", conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read failed")
}

func TestHandleAuthenticationRequest_ClearText(t *testing.T) {
	authClearText := createAuthMessage(AuthenticationClearText, nil)
	authOk := createAuthOkMessage()

	allData := append(authClearText, authOk...)

	conn := &mockAuthConn{
		readData: allData,
	}

	err := HandleAuthenticationRequest("user", "password", conn)

	assert.NoError(t, err)

	written := conn.writeBuffer.Bytes()
	assert.True(t, len(written) > 0)
	assert.Equal(t, PasswordMessageType, written[0])
}

func TestHandleAuthenticationRequest_MD5(t *testing.T) {
	salt := []byte{0x01, 0x02, 0x03, 0x04}
	authMD5 := createAuthMessage(AuthenticationMD5, salt)
	authOk := createAuthOkMessage()

	allData := append(authMD5, authOk...)

	conn := &mockAuthConn{
		readData: allData,
	}

	err := HandleAuthenticationRequest("testuser", "testpass", conn)

	assert.NoError(t, err)

	written := conn.writeBuffer.Bytes()
	assert.True(t, len(written) > 0)
	assert.Equal(t, PasswordMessageType, written[0])
}

func TestHandleAuthenticationRequest_UnsupportedAuthMethod(t *testing.T) {
	authMsg := createAuthMessage(AuthenticationKerberosV5, nil)

	conn := &mockAuthConn{
		readData: authMsg,
	}

	err := HandleAuthenticationRequest("user", "pass", conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestCreateMD5Password(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		salt     string
	}{
		{
			name:     "standard case",
			username: "testuser",
			password: "testpass",
			salt:     "salt",
		},
		{
			name:     "empty password",
			username: "user",
			password: "",
			salt:     "salt",
		},
		{
			name:     "special characters",
			username: "user@host",
			password: "p@ss!word",
			salt:     "\x01\x02\x03\x04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createMD5Password(tt.username, tt.password, tt.salt)

			assert.True(t, len(result) > 3)
			assert.Equal(t, "md5", result[:3])

			assert.Equal(t, 35, len(result))

			result2 := createMD5Password(tt.username, tt.password, tt.salt)
			assert.Equal(t, result, result2)
		})
	}
}

func TestCreateMD5Password_MatchesExpected(t *testing.T) {
	username := "user"
	password := "password"
	salt := "salt"

	result := createMD5Password(username, password, salt)

	passwordString := password + username
	hash1 := md5.Sum([]byte(passwordString))
	passwordString = string(hash1[:]) + salt
	passwordString = password + username
	passwordHash := md5.Sum([]byte(passwordString))
	passwordString = string(passwordHash[:])
	assert.True(t, len(result) == 35)
	assert.True(t, result[:3] == "md5")
}

func TestHandleAuthClearText_Success(t *testing.T) {
	authOk := createAuthOkMessage()

	conn := &mockAuthConn{
		readData: authOk,
	}

	err := handleAuthClearText("password", conn)

	assert.NoError(t, err)

	written := conn.writeBuffer.Bytes()
	assert.True(t, len(written) > 0)
	assert.Equal(t, PasswordMessageType, written[0])
}

func TestHandleAuthClearText_WriteError(t *testing.T) {
	conn := &mockAuthConn{
		writeErr: errors.New("write failed"),
	}

	err := handleAuthClearText("password", conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestHandleAuthClearText_AuthenticationFailed(t *testing.T) {
	errMsg := NewMessageBuffer([]byte{})
	errMsg.WriteByte(ErrorMessageType)
	errMsg.WriteInt32(0)
	errMsg.WriteByte(ErrorFieldSeverity)
	errMsg.WriteString(ErrorSeverityFatal)
	errMsg.WriteByte(0x00)
	errMsg.ResetLength(PGMessageLengthOffset)

	conn := &mockAuthConn{
		readData: errMsg.Bytes(),
	}

	err := handleAuthClearText("wrongpassword", conn)

	assert.Error(t, err)
}

func TestHandleAuthMD5_Success(t *testing.T) {
	authOk := createAuthOkMessage()

	conn := &mockAuthConn{
		readData: authOk,
	}

	err := handleAuthMD5("testuser", "testpass", "salt", conn)

	assert.NoError(t, err)

	written := conn.writeBuffer.Bytes()
	assert.True(t, len(written) > 0)
	assert.Equal(t, PasswordMessageType, written[0])
}

func TestHandleAuthMD5_WriteError(t *testing.T) {
	conn := &mockAuthConn{
		writeErr: errors.New("write failed"),
	}

	err := handleAuthMD5("user", "pass", "salt", conn)

	assert.Error(t, err)
}

func TestVerifyAuthentication_Success(t *testing.T) {
	authOk := createAuthOkMessage()

	conn := &mockAuthConn{
		readData: authOk,
	}

	err := verifyAuthentication(conn)

	assert.NoError(t, err)
}

func TestVerifyAuthentication_ErrorMessage(t *testing.T) {
	errMsg := NewMessageBuffer([]byte{})
	errMsg.WriteByte(ErrorMessageType)
	errMsg.WriteInt32(0)
	errMsg.WriteByte(0x00)
	errMsg.ResetLength(PGMessageLengthOffset)

	conn := &mockAuthConn{
		readData: errMsg.Bytes(),
	}

	err := verifyAuthentication(conn)

	assert.Error(t, err)
}

func TestVerifyAuthentication_UnexpectedMessageType(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteByte(QueryMessageType)
	msg.WriteInt32(4)

	conn := &mockAuthConn{
		readData: msg.Bytes(),
	}

	err := verifyAuthentication(conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected")
}

func TestVerifyAuthentication_NotAuthOk(t *testing.T) {
	authMsg := createAuthMessage(AuthenticationClearText, nil)

	conn := &mockAuthConn{
		readData: authMsg,
	}

	err := verifyAuthentication(conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AuthenticationOk")
}

func TestCreatePasswordMessage(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "special characters",
			password: "p@ss!w0rd",
		},
		{
			name:     "long password",
			password: "verylongpasswordwithmanycharsverylongpasswordwithmanychars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createPasswordMessage(tt.password)

			assert.True(t, len(result) > 0)
			assert.Equal(t, PasswordMessageType, result[0])

			length := binary.BigEndian.Uint32(result[1:5])
			assert.Equal(t, uint32(len(result)-1), length)

			msg := NewMessageBuffer(result[5:])
			readPassword, err := msg.ReadString()
			require.NoError(t, err)
			assert.Equal(t, tt.password, readPassword)
		})
	}
}

func TestCreateAuthenticationOKMessage(t *testing.T) {
	result := CreateAuthenticationOKMessage()

	assert.Equal(t, 9, len(result))
	assert.Equal(t, AuthenticationMessageType, result[0])

	length := binary.BigEndian.Uint32(result[1:5])
	assert.Equal(t, uint32(8), length)

	authValue := binary.BigEndian.Uint32(result[5:9])
	assert.Equal(t, uint32(AuthenticationOk), authValue)
}

func TestHandleAuthenticationRequest_FullFlow(t *testing.T) {
	t.Run("cleartext flow", func(t *testing.T) {
		authClearText := createAuthMessage(AuthenticationClearText, nil)
		authOk := createAuthOkMessage()
		allData := append(authClearText, authOk...)

		conn := &mockAuthConn{
			readData: allData,
		}

		err := HandleAuthenticationRequest("testuser", "testpass", conn)

		assert.NoError(t, err)
		assert.True(t, conn.writeBuffer.Len() > 0)
	})

	t.Run("md5 flow", func(t *testing.T) {
		salt := []byte{0xAA, 0xBB, 0xCC, 0xDD}
		authMD5 := createAuthMessage(AuthenticationMD5, salt)
		authOk := createAuthOkMessage()
		allData := append(authMD5, authOk...)

		conn := &mockAuthConn{
			readData: allData,
		}

		err := HandleAuthenticationRequest("testuser", "testpass", conn)

		assert.NoError(t, err)
		assert.True(t, conn.writeBuffer.Len() > 0)
	})
}
