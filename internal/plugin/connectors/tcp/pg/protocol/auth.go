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
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// HandleAuthenticationRequest sends credentials to the server and reports whether they were accepted or not.
func HandleAuthenticationRequest(username string, password string, connection net.Conn) (err error) {
	var messageType byte
	var message []byte

	if messageType, message, err = ReadMessage(connection); err != nil {
		return
	}

	if messageType == ErrorMessageType {
		err = NewError(message)
		return
	}

	if messageType != AuthenticationMessageType {
		err = fmt.Errorf("Expected %d message type, got %d", AuthenticationMessageType, messageType)
		return
	}

	var authType int32

	reader := bytes.NewReader(message)
	if err = binary.Read(reader, binary.BigEndian, &authType); err != nil {
		return
	}

	switch authType {
	case AuthenticationClearText:
		return handleAuthClearText(password, connection)
	case AuthenticationMD5:
		salt := make([]byte, 4)
		if _, err = io.ReadFull(reader, salt); err != nil {
			return
		}
		return handleAuthMD5(username, password, string(salt), connection)
	case AuthenticationOk:
		/* Covers the case where the authentication type is 'cert' or 'trust' */
		return
	}

	err = fmt.Errorf("Authentication method %d is not supported", authType)

	return
}

func createMD5Password(username string, password string, salt string) string {
	// Concatenate the password and the username together.
	passwordString := fmt.Sprintf("%s%s", password, username)

	// Compute the MD5 sum of the password+username string.
	passwordString = fmt.Sprintf("%x", md5.Sum([]byte(passwordString)))

	// Compute the MD5 sum of the password hash and the salt
	passwordString = fmt.Sprintf("%s%s", passwordString, salt)
	return fmt.Sprintf("md5%x", md5.Sum([]byte(passwordString)))
}

func handleAuthMD5(username string, password string, salt string, connection net.Conn) (err error) {
	saltedPassword := createMD5Password(username, password, salt)

	// Create the password message.
	passwordMessage := createPasswordMessage(saltedPassword)

	if _, err = connection.Write(passwordMessage); err != nil {
		return
	}

	err = verifyAuthentication(connection)
	return
}

func handleAuthClearText(password string, connection net.Conn) (err error) {
	passwordMessage := createPasswordMessage(password)

	if _, err = connection.Write(passwordMessage); err != nil {
		return
	}

	err = verifyAuthentication(connection)
	return
}

func verifyAuthentication(connection net.Conn) (err error) {
	var messageType byte
	var message []byte
	if messageType, message, err = ReadMessage(connection); err != nil {
		return
	}

	if messageType == ErrorMessageType {
		err = NewError(message)
		return
	}

	if messageType != AuthenticationMessageType {
		err = fmt.Errorf("Expected %d message type, got %d", AuthenticationMessageType, messageType)
		return
	}

	var messageValue int32
	if err = binary.Read(bytes.NewBuffer(message), binary.BigEndian, &messageValue); err != nil {
		return
	}

	if messageValue != AuthenticationOk {
		err = fmt.Errorf("Expected %d (AuthenticationOk), got %d", AuthenticationOk, messageValue)
		return
	}

	return
}

// CreatePasswordMessage creates a message which provides the password in response
// to an authentication challenge.
func createPasswordMessage(password string) []byte {
	message := NewMessageBuffer([]byte{})

	/* Set the message type */
	message.WriteByte(PasswordMessageType)

	/* Initialize the message length to zero. */
	message.WriteInt32(0)

	/* Add the password to the message. */
	message.WriteString(password)

	/* Update the message length */
	message.ResetLength(PGMessageLengthOffset)

	return message.Bytes()
}

// CreateAuthenticationOKMessage creates a Postgresql message which indicates
// successful authentication.
func CreateAuthenticationOKMessage() []byte {
	message := NewMessageBuffer([]byte{})

	message.WriteByte(AuthenticationMessageType)
	message.WriteInt32(8)
	message.WriteInt32(AuthenticationOk)

	return message.Bytes()
}
