package proxy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kgilpin/secretless-pg/conjur"
	"github.com/kgilpin/secretless-pg/connect"
	"github.com/kgilpin/secretless-pg/protocol"
)

func promptForPassword(client net.Conn) ([]byte, error) {
	message := protocol.NewMessageBuffer([]byte{})

	/* Set the message type */
	message.WriteByte(protocol.AuthenticationMessageType)

	/* Temporarily set the message length to 0. */
	message.WriteInt32(0)

	/* Set the protocol version. */
	message.WriteInt32(protocol.AuthenticationClearText)

	/* Update the message length */
	message.ResetLength(protocol.PGMessageLengthOffset)

	// Send the password message to the backend.
	_, err := connect.Send(client, message.Bytes())

	if err != nil {
		return nil, err
	}

	response := make([]byte, 4096)

	_, err = client.Read(response)
	if err != nil {
		return nil, err
	}

	message = protocol.NewMessageBuffer(response)

	code, err := message.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != protocol.PasswordMessageType {
		return nil, fmt.Errorf("Expected message %d in response to password prompt, got %d", protocol.PasswordMessageType, code)
	}

	length, err := message.ReadInt32()
	if err != nil {
		return nil, err
	}

	password, err := message.ReadBytes(int(length))
	if err != nil {
		return nil, err
	}

	password = bytes.Trim(password, "\x00")
	return password, nil
}

func authorizeWithConjur(resource, token string) error {
	allowed, err := conjur.CheckPermission(resource, token)
	if allowed {
		return nil
	} else {
		return err
	}
}

func authenticateWithPassword(password, expectedPassword string) error {
	valid := (string(password) == expectedPassword)
	if valid {
		log.Print("Password is valid")
		return nil
	} else {
		return fmt.Errorf("Password is invalid")
	}
}

/**
 * Authenticate and configure the connection to the backend.
 *
 * return abort bool, authenticationError err, err bool
 */
func (self *Handler) Authenticate() (bool, error, error) {
	var err error
	var clientPassword []byte

	// Authenticate and authorize with Conjur
	if clientPassword, err = promptForPassword(self.Client); err != nil {
		return true, nil, err
	}

	staticPassword, staticAuth := self.Config.AuthorizedUsers[self.ClientOptions.User]
	var authenticationError error
	if staticAuth {
		// There's a statically configured password
		authenticationError = authenticateWithPassword(string(clientPassword), staticPassword)
	} else {
		log.Printf("Password for '%s' not found in static configuration. Attempting Conjur authorization.", self.ClientOptions.User)

		token, err := base64.StdEncoding.DecodeString(string(clientPassword))
		if err != nil {
			return true, nil, err
		}
		authenticationError = authorizeWithConjur(self.Config.Authorization.Resource, string(token))
	}

	if authenticationError != nil {
		if self.ClientOptions.Options["application_name"] == "psql" && authenticationError == io.EOF {
			log.Printf("Got %s from psql, this is normal", err)
			return true, nil, nil
		} else {
			log.Print(authenticationError)
			var msg string
			if staticAuth {
				msg = "Login failed"
			} else {
				msg = "Conjur authorization failed"
			}
			return true, fmt.Errorf(msg), nil
		}
	}

	var backendConnection BackendConnection
	if staticAuth {
		backendConnection = StaticBackendConnection{self.Config}
	} else {
		backendConnection = ConjurBackendConnection{Resource: self.Config.Authorization.Resource}
	}
	self.BackendConnection = backendConnection

	return false, nil, nil
}
