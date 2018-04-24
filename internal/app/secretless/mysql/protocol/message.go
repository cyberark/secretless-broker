/*
 Copyright 2017 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package protocol

import (
	"bytes"
	"encoding/binary"
	"strings"
)

/* PostgreSQL message length offset constants. */
const (
	PGMessageLengthOffsetStartup int = 0
	PGMessageLengthOffset        int = 1
)

// MessageBuffer is a variable-sized byte buffer used to read and write
// PostgreSQL Frontend and Backend messages.
//
// A separate instance of a MessageBuffer should be use for reading and writing.
type MessageBuffer struct {
	buffer *bytes.Buffer
}

// NewMessageBuffer creates and intializes a new MessageBuffer using message as its
// initial contents.
func NewMessageBuffer(message []byte) *MessageBuffer {
	return &MessageBuffer{
		buffer: bytes.NewBuffer(message),
	}
}

// ReadInt32 reads an int32 from the message buffer.
//
// This function will read the next 4 available bytes from the message buffer
// and return them as an int32.
//
// panic on error.
func (message *MessageBuffer) ReadInt32() (value int32, err error) {
	err = binary.Read(message.buffer, binary.BigEndian, &value)
	return
}

// ReadByte reads a byte from the message buffer.
//
// This function will read and return the next available byte from the message
// buffer.
func (message *MessageBuffer) ReadByte() (byte, error) {
	return message.buffer.ReadByte()
}

// ReadString reads a string from the message buffer.
//
// This function will read and return the next Null terminated string from the
// message buffer.
func (message *MessageBuffer) ReadString() (string, error) {
	str, err := message.buffer.ReadString(0x00)
	return strings.Trim(str, "\x00"), err
}

// WriteByte will write the specified byte to the message buffer.
func (message *MessageBuffer) WriteByte(value byte) error {
	return message.buffer.WriteByte(value)
}

// WriteInt32 will write a 4 byte int32 to the message buffer.
func (message *MessageBuffer) WriteInt32(value int32) (err error) {
	err = binary.Write(message.buffer, binary.BigEndian, value)
	return
}

// WriteString will write a NULL terminated string to the buffer.  It is
// assumed that the incoming string has *NOT* been NULL terminated.
func (message *MessageBuffer) WriteString(value string) (int, error) {
	return message.buffer.WriteString((value + "\000"))
}

// ResetLength will reset the message length for the message.
//
// offset should be one of the PGMessageLengthOffset* constants.
func (message *MessageBuffer) ResetLength(offset int) {
	/* Get the contents of the buffer. */
	b := message.buffer.Bytes()

	/* Get the start of the message length bytes. */
	s := b[offset:]

	/* Determine the new length and set it. */
	binary.BigEndian.PutUint32(s, uint32(len(s)))
}

// Bytes gets the contents of the message buffer. This function is only
// useful after 'Write' operations as the underlying implementation will return
// the 'unread' portion of the buffer.
func (message *MessageBuffer) Bytes() []byte {
	return message.buffer.Bytes()
}
