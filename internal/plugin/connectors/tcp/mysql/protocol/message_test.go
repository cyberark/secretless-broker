package protocol

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessageBuffer(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	mb := NewMessageBuffer(data)

	assert.NotNil(t, mb)
	assert.NotNil(t, mb.buffer)
	assert.Equal(t, data, mb.Bytes())
}

func TestMessageBuffer_ReadInt32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int32
		hasError bool
	}{
		{
			name:     "valid int32",
			data:     []byte{0x00, 0x00, 0x00, 0x10},
			expected: 16,
			hasError: false,
		},
		{
			name:     "negative int32",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: -1,
			hasError: false,
		},
		{
			name:     "insufficient data",
			data:     []byte{0x01, 0x02},
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMessageBuffer(tt.data)
			result, err := mb.ReadInt32()

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageBuffer_ReadByte(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
		hasError bool
	}{
		{
			name:     "valid byte",
			data:     []byte{0x42},
			expected: 0x42,
			hasError: false,
		},
		{
			name:     "empty buffer",
			data:     []byte{},
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMessageBuffer(tt.data)
			result, err := mb.ReadByte()

			if tt.hasError {
				assert.Error(t, err)
				assert.Equal(t, io.EOF, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageBuffer_ReadString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		hasError bool
	}{
		{
			name:     "null terminated string",
			data:     []byte("hello\x00world"),
			expected: "hello",
			hasError: false,
		},
		{
			name:     "empty string",
			data:     []byte("\x00"),
			expected: "",
			hasError: false,
		},
		{
			name:     "string without null terminator",
			data:     []byte("hello"),
			expected: "hello",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMessageBuffer(tt.data)
			result, err := mb.ReadString()

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageBuffer_WriteByte(t *testing.T) {
	mb := NewMessageBuffer(nil)
	err := mb.WriteByte(0x42)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0x42}, mb.Bytes())
}

func TestMessageBuffer_WriteInt32(t *testing.T) {
	tests := []struct {
		name     string
		value    int32
		expected []byte
	}{
		{
			name:     "positive int32",
			value:    16,
			expected: []byte{0x00, 0x00, 0x00, 0x10},
		},
		{
			name:     "negative int32",
			value:    -1,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "zero",
			value:    0,
			expected: []byte{0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMessageBuffer(nil)
			err := mb.WriteInt32(tt.value)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, mb.Bytes())
		})
	}
}

func TestMessageBuffer_WriteString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []byte
	}{
		{
			name:     "normal string",
			value:    "hello",
			expected: []byte("hello\x00"),
		},
		{
			name:     "empty string",
			value:    "",
			expected: []byte("\x00"),
		},
		{
			name:     "string with special chars",
			value:    "test\nline",
			expected: []byte("test\nline\x00"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMessageBuffer(nil)
			n, err := mb.WriteString(tt.value)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), n)
			assert.Equal(t, tt.expected, mb.Bytes())
		})
	}
}

func TestMessageBuffer_ResetLength(t *testing.T) {
	mb := NewMessageBuffer(make([]byte, 20))

	mb.WriteInt32(123)
	mb.WriteString("test")

	mb.ResetLength(4)

	bytes := mb.Bytes()
	require.True(t, len(bytes) >= 8)

	lengthBytes := bytes[4:8]
	length := binary.BigEndian.Uint32(lengthBytes)
	expected := uint32(len(bytes) - 4)

	assert.Equal(t, expected, length)
}

func TestMessageBuffer_Bytes(t *testing.T) {
	initialData := []byte{0x01, 0x02, 0x03}
	mb := NewMessageBuffer(initialData)

	assert.Equal(t, initialData, mb.Bytes())

	mb.WriteByte(0x04)

	expected := []byte{0x01, 0x02, 0x03, 0x04}
	assert.Equal(t, expected, mb.Bytes())
}

func TestMessageBuffer_ReadWriteSequence(t *testing.T) {
	mb := NewMessageBuffer(nil)

	err := mb.WriteInt32(42)
	require.NoError(t, err)

	err = mb.WriteByte(0xFF)
	require.NoError(t, err)

	n, err := mb.WriteString("test")
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	readMB := NewMessageBuffer(mb.Bytes())

	intVal, err := readMB.ReadInt32()
	require.NoError(t, err)
	assert.Equal(t, int32(42), intVal)

	byteVal, err := readMB.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(0xFF), byteVal)

	strVal, err := readMB.ReadString()
	require.NoError(t, err)
	assert.Equal(t, "test", strVal)
}

func TestMessageBuffer_MultipleReads(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x10, 0xFF, 'h', 'i', '\x00'}
	mb := NewMessageBuffer(data)

	intVal, err := mb.ReadInt32()
	assert.NoError(t, err)
	assert.Equal(t, int32(16), intVal)

	byteVal, err := mb.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte(0xFF), byteVal)

	strVal, err := mb.ReadString()
	assert.NoError(t, err)
	assert.Equal(t, "hi", strVal)

	_, err = mb.ReadByte()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
}
