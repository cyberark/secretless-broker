package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNetConn struct {
	readData    []byte
	readPos     int
	writeBuffer bytes.Buffer
	readErr     error
	writeErr    error
	closed      bool
}

func (m *mockNetConn) Read(b []byte) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}
	n := copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockNetConn) Write(b []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuffer.Write(b)
}

func (m *mockNetConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr                { return nil }
func (m *mockNetConn) RemoteAddr() net.Addr               { return nil }
func (m *mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error { return nil }

func createMySQLPacket(payload []byte, seqID byte) []byte {
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(len(payload)))
	header[3] = seqID
	return append(header, payload...)
}

func TestReadPacket_ValidPacket(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	packet := createMySQLPacket(payload, 0)

	conn := &mockNetConn{readData: packet}
	result, err := ReadPacket(conn)

	assert.NoError(t, err)
	assert.Equal(t, packet, result)
}

func TestReadPacket_EmptyPayload(t *testing.T) {
	packet := createMySQLPacket([]byte{}, 0)

	conn := &mockNetConn{readData: packet}
	result, err := ReadPacket(conn)

	assert.NoError(t, err)
	assert.Equal(t, packet, result)
}

func TestReadPacket_LargePayload(t *testing.T) {
	payload := make([]byte, 1000)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	packet := createMySQLPacket(payload, 5)

	conn := &mockNetConn{readData: packet}
	result, err := ReadPacket(conn)

	assert.NoError(t, err)
	assert.Equal(t, packet, result)
}

func TestReadPacket_HeaderReadError(t *testing.T) {
	conn := &mockNetConn{
		readData: []byte{0x01, 0x02},
		readErr:  io.ErrUnexpectedEOF,
	}

	_, err := ReadPacket(conn)
	assert.Error(t, err)
}

func TestReadPacket_BodyReadError(t *testing.T) {
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, 10)
	header[3] = 0

	conn := &mockNetConn{
		readData: append(header, []byte{0x01, 0x02, 0x03, 0x04}...),
	}

	_, err := ReadPacket(conn)
	assert.Error(t, err)
}

func TestWritePacket_ValidPacket(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	packet := createMySQLPacket(payload, 0)

	conn := &mockNetConn{}
	n, err := WritePacket(packet, conn)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, conn.writeBuffer.Bytes())
}

func TestWritePacket_EmptyPacket(t *testing.T) {
	packet := createMySQLPacket([]byte{}, 0)

	conn := &mockNetConn{}
	n, err := WritePacket(packet, conn)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, conn.writeBuffer.Bytes())
}

func TestWritePacket_WriteError(t *testing.T) {
	packet := createMySQLPacket([]byte{0x01}, 0)

	conn := &mockNetConn{
		writeErr: io.ErrShortWrite,
	}

	_, err := WritePacket(packet, conn)
	assert.Error(t, err)
}

func TestProxyPacket_Success(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	packet := createMySQLPacket(payload, 0)

	src := &mockNetConn{readData: packet}
	dst := &mockNetConn{}

	result, err := ProxyPacket(src, dst)

	assert.NoError(t, err)
	assert.Equal(t, packet, result)
	assert.Equal(t, packet, dst.writeBuffer.Bytes())
}

func TestProxyPacket_ReadError(t *testing.T) {
	src := &mockNetConn{
		readData: []byte{0x01, 0x02},
		readErr:  io.ErrUnexpectedEOF,
	}
	dst := &mockNetConn{}

	_, err := ProxyPacket(src, dst)
	assert.Error(t, err)
}

func TestProxyPacket_WriteError(t *testing.T) {
	packet := createMySQLPacket([]byte{0x01}, 0)

	src := &mockNetConn{readData: packet}
	dst := &mockNetConn{writeErr: io.ErrShortWrite}

	_, err := ProxyPacket(src, dst)
	assert.Error(t, err)
}

func TestConnSettings_DeprecateEOFSet_BothSet(t *testing.T) {
	settings := &ConnSettings{
		ClientCapabilities: ClientDeprecateEOF,
		ServerCapabilities: ClientDeprecateEOF,
	}

	assert.True(t, settings.DeprecateEOFSet())
}

func TestConnSettings_DeprecateEOFSet_OnlyClientSet(t *testing.T) {
	settings := &ConnSettings{
		ClientCapabilities: ClientDeprecateEOF,
		ServerCapabilities: 0,
	}

	assert.False(t, settings.DeprecateEOFSet())
}

func TestConnSettings_DeprecateEOFSet_OnlyServerSet(t *testing.T) {
	settings := &ConnSettings{
		ClientCapabilities: 0,
		ServerCapabilities: ClientDeprecateEOF,
	}

	assert.False(t, settings.DeprecateEOFSet())
}

func TestConnSettings_DeprecateEOFSet_NeitherSet(t *testing.T) {
	settings := &ConnSettings{
		ClientCapabilities: 0,
		ServerCapabilities: 0,
	}

	assert.False(t, settings.DeprecateEOFSet())
}

func TestReadErrMessage(t *testing.T) {
	tests := []struct {
		name     string
		packet   []byte
		expected string
	}{
		{
			name:     "error with message",
			packet:   append(make([]byte, 13), []byte("Error message")...),
			expected: "Error message",
		},
		{
			name:     "empty error message",
			packet:   make([]byte, 13),
			expected: "",
		},
		{
			name:     "error with special chars",
			packet:   append(make([]byte, 13), []byte("Error: \nNew line")...),
			expected: "Error: \nNew line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReadErrMessage(tt.packet)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadResponse_OkPacket(t *testing.T) {
	okPayload := []byte{ResponseOk, 0x00, 0x00}
	okPacket := createMySQLPacket(okPayload, 1)

	conn := &mockNetConn{readData: okPacket}
	data, respType, err := ReadResponse(conn, false)

	assert.NoError(t, err)
	assert.Equal(t, byte(ResponseOk), respType)
	assert.Equal(t, okPacket, data)
}

func TestReadResponse_ErrPacket(t *testing.T) {
	errPayload := []byte{ResponseErr, 0x00, 0x00}
	errPacket := createMySQLPacket(errPayload, 1)

	conn := &mockNetConn{readData: errPacket}
	data, respType, err := ReadResponse(conn, false)

	assert.NoError(t, err)
	assert.Equal(t, byte(ResponseErr), respType)
	assert.Equal(t, errPacket, data)
}

func TestReadPrepareResponse_OkResponse(t *testing.T) {
	payload := make([]byte, 12)
	payload[0] = ResponsePrepareOk
	binary.LittleEndian.PutUint32(payload[1:5], 1)
	binary.LittleEndian.PutUint16(payload[5:7], 0)
	binary.LittleEndian.PutUint16(payload[9:11], 0)

	packet := createMySQLPacket(payload, 0)

	conn := &mockNetConn{readData: packet}
	data, respType, err := ReadPrepareResponse(conn)

	assert.NoError(t, err)
	assert.Equal(t, byte(ResponseOk), respType)
	assert.NotNil(t, data)
}

func TestReadPrepareResponse_ErrResponse(t *testing.T) {
	errPayload := []byte{ResponseErr, 0x00, 0x00}
	errPacket := createMySQLPacket(errPayload, 0)

	conn := &mockNetConn{readData: errPacket}
	data, respType, err := ReadPrepareResponse(conn)

	assert.NoError(t, err)
	assert.Equal(t, byte(ResponseErr), respType)
	assert.Equal(t, errPacket, data)
}

func TestReadShowFieldsResponse(t *testing.T) {
	okPayload := []byte{ResponseOk, 0x00}
	okPacket := createMySQLPacket(okPayload, 0)

	conn := &mockNetConn{readData: okPacket}
	data, respType, err := ReadShowFieldsResponse(conn)

	assert.NoError(t, err)
	assert.Equal(t, byte(ResponseOk), respType)
	assert.NotNil(t, data)
}

func TestReadPrepareResponse_ReadError(t *testing.T) {
	conn := &mockNetConn{
		readData: []byte{0x01},
		readErr:  io.ErrUnexpectedEOF,
	}

	_, _, err := ReadPrepareResponse(conn)
	assert.Error(t, err)
}

func TestReadPacket_SequenceOfPackets(t *testing.T) {
	packet1 := createMySQLPacket([]byte{0x01}, 0)
	packet2 := createMySQLPacket([]byte{0x02}, 1)
	packet3 := createMySQLPacket([]byte{0x03}, 2)

	allData := append(packet1, packet2...)
	allData = append(allData, packet3...)

	conn := &mockNetConn{readData: allData}

	pkt1, err := ReadPacket(conn)
	require.NoError(t, err)
	assert.Equal(t, packet1, pkt1)

	pkt2, err := ReadPacket(conn)
	require.NoError(t, err)
	assert.Equal(t, packet2, pkt2)

	pkt3, err := ReadPacket(conn)
	require.NoError(t, err)
	assert.Equal(t, packet3, pkt3)
}

func TestWritePacket_SequenceOfPackets(t *testing.T) {
	packet1 := createMySQLPacket([]byte{0x01}, 0)
	packet2 := createMySQLPacket([]byte{0x02}, 1)

	conn := &mockNetConn{}

	_, err := WritePacket(packet1, conn)
	require.NoError(t, err)

	_, err = WritePacket(packet2, conn)
	require.NoError(t, err)

	expected := append(packet1, packet2...)
	assert.Equal(t, expected, conn.writeBuffer.Bytes())
}
