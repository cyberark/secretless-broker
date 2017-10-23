package proxy

import (
	"fmt"
	"log"

	"github.com/kgilpin/secretless-pg/connect"
	"github.com/kgilpin/secretless-pg/protocol"
)

/*
 * Parse client options from the buffer and store the required parameters.
 * If an error is reported here, it should be propagated as Fatal to the client.
 */
func (self *ClientOptions) Parse(message *protocol.MessageBuffer) error {
	self.Options = make(map[string]string)
	for {
		param, err := message.ReadString()
		value, err := message.ReadString()
		if err != nil || param == "\x00" {
			break
		}

		self.Options[param] = value
	}

	log.Printf("Client options : %s", self.Options)

	var ok bool

	self.User, ok = self.Options["user"]
	if !ok {
		return fmt.Errorf("No 'user' found in connect options")
	}
	self.Database, ok = self.Options["database"]
	if !ok {
		return fmt.Errorf("No 'database' found in connect options")
	}

	return nil
}

/*
 * Perform the startup handshake with the client and obtain the client options.
 * If an error is reported here, it should be propagated as Fatal to the client.
 */
func (self *Handler) Startup() error {
	log.Printf("Handling connection %v", self.Client)

	/* Get the self.Client startup message. */
	message, length, err := connect.Receive(self.Client)
	if err != nil {
		return fmt.Errorf("Error receiving startup message from self.Client: %s", err)
	}

	/* Get the protocol from the startup message.*/
	version := protocol.GetVersion(message)

	log.Printf("self.Client version : %v, (SSL mode: %v)", version, version == protocol.SSLRequestCode)

	/* Handle the case where the startup message was an SSL request. */
	if version == protocol.SSLRequestCode {
		return fmt.Errorf("SSL not supported")
	}

	/* Now read the startup parameters */
	startup := protocol.NewMessageBuffer(message[8:length])

	self.ClientOptions = ClientOptions{}
	return self.ClientOptions.Parse(startup)
}
