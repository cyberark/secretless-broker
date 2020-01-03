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

// ErrorContainer interface makes it possible
// to have Go errors that can contain rich protocol specific information
// and have the smarts to encode themselves into a MYSQL error packet
type ErrorContainer interface {
	GetPacket() []byte
}

// Error is a MySQL processing error.
type Error struct {
	Code     uint16
	SQLState string
	Message  string
}

// NewGenericError returns a MySQL protocol specific error, to handle cases
// that don't fit neatly into the MySQL code categories.  This makes sense,
// since a Go error, eg, is not a true "protocol" error but instead something
// resulting from Secretless's role as a proxy, which MySQL is unaware of.
//
// That said, we should try to use MySQL specific error codes wherever we
// can, so that client error messages will be more meaningful.
//
// TODO: Replace instances of generic error with specific MySQL error codes.
func NewGenericError(goErr error) Error {
	return Error{
		Code:     CRUnknownError,
		SQLState: ErrorCodeInternalError,
		Message:  goErr.Error(),
	}
}

// ErrNoTLS is a MySQL protocol error raised when SSL is required but the
// server doesn't support it.
var ErrNoTLS = Error{
	Code:     CRSSLConnectionError,
	SQLState: ErrorCodeInternalError,
	Message:  "SSL connection error: SSL is required but the server doesn't support it",
}

func (e Error) Error() string {
	return fmt.Sprintf("ERROR: %v (%s): %s", e.Code, e.SQLState, e.Message)
}

// GetPacket formats an Error into a protocol message.
// https://dev.mysql.com/doc/internals/en/packet-ERR_Packet.html
func (e Error) GetPacket() []byte {
	data := make([]byte, 4, 4+1+2+1+5+len(e.Message))
	data = append(data, 0xff)
	data = append(data, byte(e.Code), byte(e.Code>>8))

	// TODO: for client41
	data = append(data, '#')
	data = append(data, e.SQLState...)
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
