package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// SingleUseConnector creates an authenticated connection to a target TCP service.
type SingleUseConnector struct {
	logger log.Logger
}

// Connect is the function that implements the tcp.Connector func signature in this
// example plugin. It has access to the client connection and the credentials (as a map),
// and is expected to return the target service connection.
//
// This example connector works as follows:
// 1. Waits for the initial message from the client
// 2. Connect to a target service whose address is the value of the credential identified
// by the key "address"
// 3. Inject credentials from a credential identified by the key "auth"
// 4. Write the initial message from the client with some modification
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {

	connector.logger.Debugln("Waiting for initial write from client")
	clientInitMsg, _, err := bufio.NewReader(clientConn).ReadLine()
	if err != nil {
		return nil, err
	}

	connector.logger.Debugln("Dialing target service")
	conn, err := net.Dial("tcp", string(credentialValuesByID["address"]))
	if err != nil {
		return nil, err
	}

	connector.logger.Debugln("Sending packet with injected credentials to target service")
	credInjectionPacket := []byte(
		fmt.Sprintf(
			"credential injection: %s\n",
			string(credentialValuesByID["auth"]),
		),
	)
	_, err = conn.Write(credInjectionPacket)
	if err != nil {
		return nil, err
	}

	connector.logger.Debugln("Sending modified client initial packet to target service")
	initMsgPacket := []byte(
		fmt.Sprintf(
			"initial message from client: %s\n",
			string(clientInitMsg),
		),
	)
	_, err = conn.Write(initMsgPacket)
	if err != nil {
		return nil, err
	}

	connector.logger.Debugln("Successfully connected to target service")
	return conn, nil
}
