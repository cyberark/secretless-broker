package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name: "fatal error",
			err: Error{
				Severity: ErrorSeverityFatal,
				Message:  "connection failed",
			},
			expected: "pg: FATAL: connection failed",
		},
		{
			name: "panic error",
			err: Error{
				Severity: ErrorSeverityPanic,
				Message:  "system crash",
			},
			expected: "pg: PANIC: system crash",
		},
		{
			name: "warning",
			err: Error{
				Severity: ErrorSeverityWarning,
				Message:  "deprecated feature",
			},
			expected: "pg: WARNING: deprecated feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewError_Success(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteByte(ErrorFieldSeverity)
	msg.WriteString(ErrorSeverityFatal)
	msg.WriteByte(ErrorFieldCode)
	msg.WriteString("28P01")
	msg.WriteByte(ErrorFieldMessage)
	msg.WriteString("authentication failed")
	msg.WriteByte(ErrorFieldMessageDetail)
	msg.WriteString("password is incorrect")
	msg.WriteByte(ErrorFieldMessageHint)
	msg.WriteString("check your password")
	msg.WriteByte(0x00)

	result := NewError(msg.Bytes())

	require.NotNil(t, result)
	pgErr, ok := result.(*Error)
	require.True(t, ok)
	assert.Equal(t, ErrorSeverityFatal, pgErr.Severity)
	assert.Equal(t, "28P01", pgErr.Code)
	assert.Equal(t, "authentication failed", pgErr.Message)
	assert.Equal(t, "password is incorrect", pgErr.Detail)
	assert.Equal(t, "check your password", pgErr.Hint)
}

func TestNewError_MinimalFields(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteByte(ErrorFieldSeverity)
	msg.WriteString(ErrorSeverityFatal)
	msg.WriteByte(ErrorFieldCode)
	msg.WriteString(ErrorCodeInternalError)
	msg.WriteByte(ErrorFieldMessage)
	msg.WriteString("internal error")
	msg.WriteByte(0x00)

	result := NewError(msg.Bytes())

	require.NotNil(t, result)
	pgErr, ok := result.(*Error)
	require.True(t, ok)
	assert.Equal(t, ErrorSeverityFatal, pgErr.Severity)
	assert.Equal(t, ErrorCodeInternalError, pgErr.Code)
	assert.Equal(t, "internal error", pgErr.Message)
	assert.Empty(t, pgErr.Detail)
	assert.Empty(t, pgErr.Hint)
}

func TestError_GetPacket_PacketStructure(t *testing.T) {
	err := &Error{
		Severity: ErrorSeverityFatal,
		Code:     ErrorCodeInternalError,
		Message:  "test",
	}

	packet := err.GetPacket()

	assert.Equal(t, ErrorMessageType, packet[0])
	assert.Equal(t, byte(0x00), packet[len(packet)-1])

	msg := NewMessageBuffer(packet[5:])

	fieldType, readErr := msg.ReadByte()
	require.NoError(t, readErr)
	assert.Equal(t, ErrorFieldSeverity, fieldType)

	severity, readErr := msg.ReadString()
	require.NoError(t, readErr)
	assert.Equal(t, ErrorSeverityFatal, severity)
}
