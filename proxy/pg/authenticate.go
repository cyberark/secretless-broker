package pg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kgilpin/secretless/conjur"
	"github.com/kgilpin/secretless/connect"
	"github.com/kgilpin/secretless/protocol"
)

/**
 * Authenticate and authorize a client.
 */
type Authorizer interface {
	/**
	* Return authenticationError, otherError
	*/
	Authorize(clientPassword []byte) (error, error)
}

type SuccessAuthorizer struct {
}

type StaticAuthorizer struct {
	User string
	AuthorizedUsers map[string]string
}

type ConjurAuthorizer struct {
	Resource string
}

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

func (self SuccessAuthorizer) Authorize(clientPassword []byte) (error, error) {
	return nil, nil
}

func (self StaticAuthorizer) Authorize(clientPassword []byte) (error, error) {
	log.Printf("Authenticating %s from static password configuration", self.User)

	expectedPassword := self.AuthorizedUsers[self.User]

	valid := expectedPassword != "" && (string(clientPassword) == expectedPassword)
	if valid {
		log.Print("Password is valid")
		return nil, nil
	} else {
		return fmt.Errorf("Login failed"), nil
	}
}

func (self ConjurAuthorizer) Authorize(clientPassword []byte) (error, error) {
	token, err := base64.StdEncoding.DecodeString(string(clientPassword))
	if err != nil {
		return nil, err
	}
	allowed, err := conjur.CheckPermission(self.Resource, conjur.AccessToken{Token: string(token)})
	if err != nil {
		return nil, err
	}	
	if allowed {
		return nil, nil
	} else {
		return fmt.Errorf("Conjur authorization failed"), nil
	}
}

/**
 * Authenticate and configure the connection to the backend.
 *
 * return abort bool, authenticationError err, err bool
 */
func (self *PGHandler) Authenticate() (bool, error, error) {
	var authenticationError, err error
	var clientPassword []byte
	var authorizer Authorizer

	if self.Config.Authorization.None {
		authorizer = SuccessAuthorizer{}
	} else {
		if clientPassword, err = promptForPassword(self.Client); err != nil {
			if self.ClientOptions.Options["application_name"] == "psql" && err == io.EOF {
				log.Printf("Got %s from psql, this is normal", err)
				return true, nil, nil
			} else {
				return true, nil, err
			}
		}

		if self.Config.Authorization.Conjur != "" {
			authorizer = ConjurAuthorizer{self.Config.Authorization.Conjur}
		} else {
			authorizer = StaticAuthorizer{self.ClientOptions.User, self.Config.Authorization.Passwords}
		}
	}

	authenticationError, err = authorizer.Authorize(clientPassword)
	if err != nil {
		return true, nil, err
	} else if authenticationError != nil {
		return true, authenticationError, nil
	}

	return false, nil, nil
}
