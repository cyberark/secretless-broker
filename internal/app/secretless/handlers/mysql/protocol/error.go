package protocol

import (
	"bytes"
	"fmt"
)

/* MySQL Error Codes */
const (
	CRUnknownError  = "2000"
	malformedPacket = "2027"
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
func (e *Error) GetMessage() []byte {
	msg := NewMessageBuffer([]byte{})

	msg.WriteString("Error: ")
	msg.WriteString(e.Code)

	msg.WriteString("SQLSTATE: ")
	msg.WriteString(e.SQLSTATE)

	msg.WriteString("Message: ")
	msg.WriteString(e.Message)

	msg.WriteByte(0x00) // null terminate the message

	return msg.Bytes()
}

// ParseError takes in stream and returns error
func ParseError(data []byte) (e *Error) {
	e = &Error{}

	buf := bytes.NewBuffer(data)
	if _, err := buf.ReadByte(); err != nil {
		return &Error{malformedPacket, "CR_MALFORMED_PACKET", "Malformed packet"}
	}

	// read error code
	codeBuf := make([]byte, 2)
	if _, err := buf.Read(codeBuf); err != nil {
		return &Error{malformedPacket, "CR_MALFORMED_PACKET", "Malformed packet"}
	}
	e.Code = string(codeBuf)

	// read sql state
	if _, err := buf.ReadByte(); err != nil {
		return &Error{malformedPacket, "CR_MALFORMED_PACKET", "Malformed packet"}
	}

	sqlStateBuf := make([]byte, 5)
	if _, err := buf.Read(sqlStateBuf); err != nil {
		return &Error{malformedPacket, "CR_MALFORMED_PACKET", "Malformed packet"}
	}
	e.SQLSTATE = string(sqlStateBuf)

	// read error message
	messageBuf := make([]byte, buf.Len())
	if _, err := buf.Read(messageBuf); err != nil {
		return &Error{malformedPacket, "CR_MALFORMED_PACKET", "Malformed packet"}
	}
	e.Message = string(messageBuf)

	return
}
