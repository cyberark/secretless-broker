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
	"net"
)

// HandleAuthenticationRequest sends credentials to the server and reports whether they were accepted or not.
func HandleAuthenticationRequest(username string, password string, connection net.Conn) (err error) {
	return
}

func createMD5Password(username string, password string, salt string) string {
	return ""
}

func handleAuthMD5(username string, password string, salt string, connection net.Conn) (err error) {
	return
}

func handleAuthClearText(password string, connection net.Conn) (err error) {
	return
}

func verifyAuthentication(connection net.Conn) (err error) {
	return
}

// CreatePasswordMessage creates a message which provides the password in response
// to an authentication challenge.
func createPasswordMessage(password string) []byte {
	return []byte{}
}

// CreateAuthenticationOKMessage creates a Postgresql message which indicates
// successful authentication.
func CreateAuthenticationOKMessage() []byte {
	return []byte{}
}
