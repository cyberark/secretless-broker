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

// ParseStartupMessage parses the pg startup message.
func ParseStartupMessage(message []byte) (version int32, options map[string]string, err error) {
	messageBuffer := NewMessageBuffer(message)
	if version, err = messageBuffer.ReadInt32(); err != nil {
		return
	}

	options = make(map[string]string)
	for {
		param, err := messageBuffer.ReadString()
		if err != nil || param == "\x00" {
			break
		}
		value, err := messageBuffer.ReadString()
		if err != nil || value == "\x00" {
			break
		}

		options[param] = value
	}

	return
}

// CreateStartupMessage creates a PG startup message. This message is used to
// startup all connections with a PG backend.
func CreateStartupMessage(
	version int32,
	username string,
	database string,
	options map[string]string,
) []byte {
	message := NewMessageBuffer([]byte{})

	/* Temporarily set the message length to 0. */
	message.WriteInt32(0)

	/* Set the protocol version. */
	message.WriteInt32(version)

	/*
	 * The protocol version number is followed by one or more pairs of
	 * parameter name and value strings. A zero byte is required as a
	 * terminator after the last name/value pair. Parameters can appear in any
	 * order. 'user' is required, others are optional.
	 */

	/* Set the 'user' parameter.  This is the only *required* parameter. */
	message.WriteString("user")
	message.WriteString(username)

	/*
	 * Set the 'database' parameter.  If no database name has been specified,
	 * then the default value is the user's name.
	 */
	message.WriteString("database")
	message.WriteString(database)

	/* Set the remaining options as specified. */
	for option, value := range options {
		message.WriteString(option)
		message.WriteString(value)
	}

	/* The message should end with a NULL byte. */
	message.WriteByte(0x00)

	/* update the msg len */
	message.ResetLength(PGMessageLengthOffsetStartup)

	return message.Bytes()
}
