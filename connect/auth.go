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

package connect

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/kgilpin/secretless-pg/protocol"
)

/*
 * Handle authentication requests that are sent by the backend to the client.
 *
 * connection - the connection to authenticate against.
 * message - the authentication message sent by the backend.
 */
func HandleAuthenticationRequest(username string, password string, connection net.Conn, message []byte) ([]byte, bool) {
	var authType int32

	// Read message length.
	// msgLength := protocol.GetMessageLength(message)

	// Read authentication type.
	reader := bytes.NewReader(message[5:9])
	binary.Read(reader, binary.BigEndian, &authType)

	switch authType {
	case protocol.AuthenticationKerberosV5:
		log.Print("KerberosV5 authentication is not currently supported.")
	case protocol.AuthenticationClearText:
		log.Print("Authenticating with clear text password.")
		return handleAuthClearText(password, connection)
	case protocol.AuthenticationMD5:
		log.Print("Authenticating with MD5 password.")
		return handleAuthMD5(username, password, connection, message)
	case protocol.AuthenticationSCM:
		log.Print("SCM authentication is not currently supported.")
	case protocol.AuthenticationGSS:
		log.Print("GSS authentication is not currently supported.")
	case protocol.AuthenticationGSSContinue:
		log.Print("GSS authentication is not currently supported.")
	case protocol.AuthenticationSSPI:
		log.Print("SSPI authentication is not currently supported.")
	case protocol.AuthenticationOk:
		/* Covers the case where the authentication type is 'cert' or 'trust' */
		return make([]byte, 0), true
	default:
		log.Printf("Unknown authentication method: %d", authType)
	}

	return nil, false
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

func handleAuthMD5(username string, password string, connection net.Conn, message []byte) ([]byte, bool) {
	salt := string(message[9:13])

	password = createMD5Password(username, password, salt)

	// Create the password message.
	passwordMessage := protocol.CreatePasswordMessage(password)

	// Send the password message to the backend.
	_, err := Send(connection, passwordMessage)

	// Check that write was successful.
	if err != nil {
		log.Print("Error sending password message to the backend.")
		log.Printf("Error: %s", err.Error())
	}

	// Read response from password message.
	message, length, err := Receive(connection)

	// Check that read was successful.
	if err != nil {
		log.Print("Error receiving authentication response from the backend.")
		log.Printf("Error: %s", err.Error())
	}

	return message[:length], protocol.IsAuthenticationOk(message)
}

func handleAuthClearText(password string, connection net.Conn) ([]byte, bool) {
	passwordMessage := protocol.CreatePasswordMessage(password)

	_, err := connection.Write(passwordMessage)

	if err != nil {
		log.Print("Error sending clear text password message to the backend.")
		log.Printf("Error: %s", err.Error())
	}

	message, length, err := Receive(connection)

	if err != nil {
		log.Print("Error receiving clear text authentication response.")
		log.Printf("Error: %s", err.Error())
	}

	return message[:length], protocol.IsAuthenticationOk(message)
}
