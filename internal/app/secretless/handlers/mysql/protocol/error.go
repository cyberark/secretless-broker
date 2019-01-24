package protocol

import (
	"fmt"
)

/* MySQL Error Codes */
const (
	CRUnknownError  = 2000
	malformedPacket = 2027
	// CRSSLConnectionError is CR_SSL_CONNECTION_ERROR
	CRSSLConnectionError = 2026
)

const (
	// ErrorCodeInternalError indicates an unspecified internal error.
	ErrorCodeInternalError = "HY000"
)

// Error is a MySQL processing error.
type Error struct {
	Code        uint16
	SQLSTATE    string
	Message     string
}

func (e *Error) Error() string {
	return fmt.Sprintf("ERROR: %s (%s): %s", e.Code, e.SQLSTATE, e.Message)
}

const ERR_HEADER = 0xff
// GetMessage formats an Error into a protocol message.
// https://dev.mysql.com/doc/internals/en/packet-ERR_Packet.html
func (e *Error) GetMessage() []byte {
	data := make([]byte, 4, 4 + 1 + 2 + 1 + 5 + len(e.Message))
	data = append(data, ERR_HEADER)
	data = append(data, byte(e.Code), byte(e.Code>>8))

	// TODO: for client41
	data = append(data, '#')
	data = append(data, e.SQLSTATE...)

	data = append(data, e.Message...)

	// prepare message
	length := len(data) - 4
	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	//sequenceID defaults to 0
	//expected to be overwritten by writer
	data[3] = byte(0)

	return data
}
