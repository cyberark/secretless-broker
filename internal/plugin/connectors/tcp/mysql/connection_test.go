package mysql

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientConnection(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := NewClientConnection(mockConn)

	assert.NotNil(t, conn)
	assert.Equal(t, mockConn, conn.conn)
	assert.Equal(t, byte(0), conn.sequenceID)
}

func TestNewBackendConnection(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := NewBackendConnection(mockConn)

	assert.NotNil(t, conn)
	assert.Equal(t, mockConn, conn.conn)
	assert.Equal(t, byte(1), conn.sequenceID)
}

func TestConnection_RawConnection(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &Connection{conn: mockConn}

	result := conn.RawConnection()

	assert.Equal(t, mockConn, result)
}

func TestConnection_write_Success(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	packet := createTestPacket([]byte{0x01, 0x02, 0x03}, 0)

	err := conn.write(packet)

	assert.NoError(t, err)
	assert.Equal(t, byte(1), conn.sequenceID, "sequence ID should increment")
	assert.True(t, len(mockConn.writeBuffer) > 0, "data should be written")
}

func TestConnection_write_SequenceIDIncrement(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 5,
	}

	packet1 := createTestPacket([]byte{0x01}, 0)
	packet2 := createTestPacket([]byte{0x02}, 0)
	packet3 := createTestPacket([]byte{0x03}, 0)

	err := conn.write(packet1)
	require.NoError(t, err)
	assert.Equal(t, byte(6), conn.sequenceID)

	err = conn.write(packet2)
	require.NoError(t, err)
	assert.Equal(t, byte(7), conn.sequenceID)

	err = conn.write(packet3)
	require.NoError(t, err)
	assert.Equal(t, byte(8), conn.sequenceID)
}

func TestConnection_write_Error(t *testing.T) {
	mockConn := &mockNetConn{
		writeErr: errors.New("write error"),
	}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	packet := createTestPacket([]byte{0x01}, 0)

	err := conn.write(packet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

func TestConnection_read_Success(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	packet := createMySQLPacket(payload, 2)

	mockConn := &mockNetConn{
		readData: packet,
	}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	result, err := conn.read()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, byte(3), conn.sequenceID, "sequence ID should be read from packet and incremented")
}

func TestConnection_read_SequenceIDUpdate(t *testing.T) {
	tests := []struct {
		name          string
		packetSeqID   byte
		initialSeqID  byte
		expectedSeqID byte
	}{
		{
			name:          "sequence ID 0",
			packetSeqID:   0,
			initialSeqID:  5,
			expectedSeqID: 1,
		},
		{
			name:          "sequence ID 10",
			packetSeqID:   10,
			initialSeqID:  0,
			expectedSeqID: 11,
		},
		{
			name:          "sequence ID 255 wraps to 0",
			packetSeqID:   255,
			initialSeqID:  0,
			expectedSeqID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := createMySQLPacket([]byte{0x01}, tt.packetSeqID)
			mockConn := &mockNetConn{readData: packet}
			conn := &Connection{
				conn:       mockConn,
				sequenceID: tt.initialSeqID,
			}

			_, err := conn.read()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSeqID, conn.sequenceID)
		})
	}
}

func TestConnection_read_Error(t *testing.T) {
	mockConn := &mockNetConn{
		readErr: errors.New("read error"),
	}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	result, err := conn.read()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "read error")
}

func TestConnection_read_EOF(t *testing.T) {
	mockConn := &mockNetConn{
		readData: []byte{},
	}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	result, err := conn.read()

	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, result)
}

func TestConnection_WriteReadSequence(t *testing.T) {
	writeConn := &mockNetConn{}
	conn := &Connection{
		conn:       writeConn,
		sequenceID: 0,
	}

	packet1 := createTestPacket([]byte{0x01}, 0)
	packet2 := createTestPacket([]byte{0x02}, 0)
	packet3 := createTestPacket([]byte{0x03}, 0)

	err := conn.write(packet1)
	require.NoError(t, err)
	assert.Equal(t, byte(1), conn.sequenceID)

	err = conn.write(packet2)
	require.NoError(t, err)
	assert.Equal(t, byte(2), conn.sequenceID)

	err = conn.write(packet3)
	require.NoError(t, err)
	assert.Equal(t, byte(3), conn.sequenceID)

	readPacket := createMySQLPacket([]byte{0xFF}, 3)
	readConn := &mockNetConn{readData: readPacket}
	conn.SetConnection(readConn)

	result, err := conn.read()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, byte(4), conn.sequenceID)
}

func TestConnection_SequenceIDWrapAround(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 254,
	}

	packet1 := createTestPacket([]byte{0x01}, 0)
	packet2 := createTestPacket([]byte{0x02}, 0)

	err := conn.write(packet1)
	require.NoError(t, err)
	assert.Equal(t, byte(255), conn.sequenceID)

	err = conn.write(packet2)
	require.NoError(t, err)
	assert.Equal(t, byte(0), conn.sequenceID, "sequence ID should wrap around to 0")
}

func TestConnection_MultipleReads(t *testing.T) {
	packet1 := createMySQLPacket([]byte{0x01}, 0)
	packet2 := createMySQLPacket([]byte{0x02}, 1)
	packet3 := createMySQLPacket([]byte{0x03}, 2)

	allData := append(packet1, packet2...)
	allData = append(allData, packet3...)

	mockConn := &mockNetConn{readData: allData}
	conn := &Connection{
		conn:       mockConn,
		sequenceID: 0,
	}

	result1, err := conn.read()
	require.NoError(t, err)
	assert.NotNil(t, result1)
	assert.Equal(t, byte(1), conn.sequenceID)

	result2, err := conn.read()
	require.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, byte(2), conn.sequenceID)

	result3, err := conn.read()
	require.NoError(t, err)
	assert.NotNil(t, result3)
	assert.Equal(t, byte(3), conn.sequenceID)
}

func TestConnection_RawConnectionIntegration(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := NewClientConnection(mockConn)

	raw := conn.RawConnection()
	assert.Equal(t, mockConn, raw)

	packet := createTestPacket([]byte{0x01, 0x02}, 0)
	err := conn.write(packet)
	require.NoError(t, err)

	assert.True(t, len(mockConn.writeBuffer) > 0)
}

func createTestPacket(payload []byte, seqID byte) Packet {
	packet := make([]byte, 4+len(payload))
	packet[0] = byte(len(payload))
	packet[1] = byte(len(payload) >> 8)
	packet[2] = byte(len(payload) >> 16)
	packet[3] = seqID
	copy(packet[4:], payload)
	return packet
}

func createMySQLPacket(payload []byte, seqID byte) []byte {
	packet := make([]byte, 4+len(payload))
	packet[0] = byte(len(payload))
	packet[1] = byte(len(payload) >> 8)
	packet[2] = byte(len(payload) >> 16)
	packet[3] = seqID
	copy(packet[4:], payload)
	return packet
}
