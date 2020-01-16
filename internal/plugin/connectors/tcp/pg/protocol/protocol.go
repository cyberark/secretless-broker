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
	"encoding/binary"
	"errors"
	"io"
)

/* PostgreSQL Protocol Version/Code constants */
const (
	ProtocolVersion int32 = 196608
	SSLRequestCode  int32 = 80877103

	/* SSL Responses */
	SSLAllowed    byte = 'S'
	SSLNotAllowed byte = 'N'
)

/* PostgreSQL Message Type constants. */
const (
	AuthenticationMessageType  byte = 'R'
	ErrorMessageType           byte = 'E'
	EmptyQueryMessageType      byte = 'I'
	DescribeMessageType        byte = 'D'
	RowDescriptionMessageType  byte = 'T'
	DataRowMessageType         byte = 'D'
	QueryMessageType           byte = 'Q'
	CommandCompleteMessageType byte = 'C'
	TerminateMessageType       byte = 'X'
	NoticeMessageType          byte = 'N'
	PasswordMessageType        byte = 'p'
	ReadyForQueryMessageType   byte = 'Z'
)

/* PostgreSQL Authentication Method constants. */
const (
	AuthenticationOk          int32 = 0
	AuthenticationKerberosV5  int32 = 2
	AuthenticationClearText   int32 = 3
	AuthenticationMD5         int32 = 5
	AuthenticationSCM         int32 = 6
	AuthenticationGSS         int32 = 7
	AuthenticationGSSContinue int32 = 8
	AuthenticationSSPI        int32 = 9
)

// ReadStartupMessage reads the startup message. The startup message is the same as a regular
// message except it does not begin with a message type byte.
func ReadStartupMessage(client io.Reader) ([]byte, error) {
	return readMessage(client)
}

// ReadMessage accepts an incoming message. The first byte is the message type, the second int32
// is the message length, and the rest of the bytes are the message body.
func ReadMessage(client io.Reader) (messageType byte, message []byte, err error) {
	messageTypeBytes := make([]byte, 1)
	if err = binary.Read(client, binary.BigEndian, &messageTypeBytes); err != nil {
		return
	}
	messageType = messageTypeBytes[0]

	message, err = readMessage(client)

	return
}

func readMessage(client io.Reader) (message []byte, err error) {
	var messageLength int32

	if err = binary.Read(client, binary.BigEndian, &messageLength); err != nil {
		return
	}

	if messageLength < 4 {
		err = errors.New("invalid message length < 4")
		return
	}
	// Build a buffer of the appropriate size and fill it
	message = make([]byte, messageLength-4)
	if _, err = io.ReadFull(client, message); err != nil {
		return
	}

	return
}
