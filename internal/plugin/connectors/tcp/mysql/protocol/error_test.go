package protocol

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenericError(t *testing.T) {
	goErr := errors.New("test error message")
	result := NewGenericError(goErr)

	assert.Equal(t, uint16(CRUnknownError), result.Code)
	assert.Equal(t, ErrorCodeInternalError, result.SQLState)
	assert.Equal(t, "test error message", result.Message)
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name: "generic error",
			err: Error{
				Code:     CRUnknownError,
				SQLState: ErrorCodeInternalError,
				Message:  "something went wrong",
			},
			expected: "ERROR: 2000 (HY000): something went wrong",
		},
		{
			name: "SSL error",
			err: Error{
				Code:     CRSSLConnectionError,
				SQLState: ErrorCodeInternalError,
				Message:  "SSL connection failed",
			},
			expected: "ERROR: 2026 (HY000): SSL connection failed",
		},
		{
			name: "malformed packet error",
			err: Error{
				Code:     malformedPacket,
				SQLState: ErrorCodeInternalError,
				Message:  "packet is malformed",
			},
			expected: "ERROR: 2027 (HY000): packet is malformed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrNoTLS_Properties(t *testing.T) {
	assert.Equal(t, uint16(CRSSLConnectionError), ErrNoTLS.Code)
	assert.Equal(t, ErrorCodeInternalError, ErrNoTLS.SQLState)
	assert.Equal(t, "SSL connection error: SSL is required but the server doesn't support it", ErrNoTLS.Message)
}

func TestErrNoTLS_Error(t *testing.T) {
	expected := "ERROR: 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it"
	assert.Equal(t, expected, ErrNoTLS.Error())
}

func TestError_GetPacket(t *testing.T) {
	err := Error{
		Code:     CRUnknownError,
		SQLState: ErrorCodeInternalError,
		Message:  "test",
	}

	packet := err.GetPacket()

	assert.True(t, len(packet) >= 13, "packet should be at least 13 bytes")
	assert.Equal(t, byte(0xff), packet[4], "byte 4 should be error indicator (0xff)")
	errorCode := uint16(packet[5]) | uint16(packet[6])<<8
	assert.Equal(t, uint16(CRUnknownError), errorCode)
	assert.Equal(t, byte('#'), packet[7], "byte 7 should be SQL state marker (#)")
	sqlState := string(packet[8:13])
	assert.Equal(t, ErrorCodeInternalError, sqlState)
	message := string(packet[13:])
	assert.Equal(t, "test", message)
}

func TestError_GetPacket_Length(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "short message",
			message: "err",
		},
		{
			name:    "medium message",
			message: "this is a test error message",
		},
		{
			name:    "long message",
			message: "this is a very long error message that contains many characters to test packet length calculation",
		},
		{
			name:    "empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error{
				Code:     CRUnknownError,
				SQLState: ErrorCodeInternalError,
				Message:  tt.message,
			}

			packet := err.GetPacket()
			packetLength := int(packet[0]) | int(packet[1])<<8 | int(packet[2])<<16
			actualLength := len(packet) - 4

			assert.Equal(t, actualLength, packetLength, "packet length should match actual payload length")
		})
	}
}

func TestError_GetPacket_SequenceID(t *testing.T) {
	err := Error{
		Code:     CRUnknownError,
		SQLState: ErrorCodeInternalError,
		Message:  "test",
	}

	packet := err.GetPacket()
	assert.Equal(t, byte(0), packet[3], "sequence ID should default to 0")
}

func TestError_GetPacket_DifferentCodes(t *testing.T) {
	tests := []struct {
		name string
		code uint16
	}{
		{
			name: "CR_UNKNOWN_ERROR",
			code: CRUnknownError,
		},
		{
			name: "CR_SSL_CONNECTION_ERROR",
			code: CRSSLConnectionError,
		},
		{
			name: "malformed_packet",
			code: malformedPacket,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error{
				Code:     tt.code,
				SQLState: ErrorCodeInternalError,
				Message:  "test",
			}

			packet := err.GetPacket()
			errorCode := uint16(packet[5]) | uint16(packet[6])<<8

			assert.Equal(t, tt.code, errorCode)
		})
	}
}

func TestError_Implements_ErrorContainer(t *testing.T) {
	var _ ErrorContainer = Error{}
	var _ ErrorContainer = &Error{}
}

func TestError_Implements_GoError(t *testing.T) {
	var _ error = Error{}
	var _ error = &Error{}
}
