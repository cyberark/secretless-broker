package protocol

import (
	"fmt"
)

/* MySQL Error Codes */
const (
	CRUnknownError = "2000"
)

const (
	// ErrorCodeInternalError indicates an unspecified internal error.
	ErrorCodeInternalError = "HY000"
)

// Error is a MySQL processing error.
type Error struct {
	Code     string
	SQLSTATE string
	Message  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("ERROR: %s (%s): %s", e.Code, e.SQLSTATE, e.Message)
}

// GetMessage formats an Error into a protocol message.
// TODO update for MySQL
func (e *Error) GetMessage() []byte {
	msg := NewMessageBuffer([]byte{})

	msg.WriteString("Error: ")
	msg.WriteString(e.Code)

	msg.WriteString("SQLSTATE: ")
	msg.WriteString(e.SQLSTATE)

	msg.WriteString("Message: ")
	msg.WriteString(e.Message)

	msg.WriteByte(0x00) // null terminate the message

	//msg.ResetLength(PGMessageLengthOffset)

	return msg.Bytes()
}
