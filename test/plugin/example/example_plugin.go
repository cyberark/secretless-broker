package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// Connector creates an authenticated connection to a target TCP service.
//
type Connector struct{
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
func (c *Connector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	c.logger.Debugln("waiting for initial write from client")
	clientInitMsg, _, err := bufio.NewReader(clientConn).ReadLine()
	if err != nil {
		return nil, err
	}


	c.logger.Debugln("dialing target service")
	conn, err := net.Dial("tcp", string(credentialValuesByID["address"]))
	if err != nil {
		return nil, err
	}

	c.logger.Debugln("sending packet with injected credentials to target service")
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

	c.logger.Debugln("sending modified client initial packet to target service")
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

	c.logger.Debugln("successfully connected to target service")
	return conn, nil
}

// NewConnector is a required method on the tcp.Plugin interface. It returns a
// tcp.Connector.
//
// The single argument passed in is of type connector.Resources. It contains
// connector-specific config and a logger.
func NewConnector(conRes connector.Resources) tcp.Connector {
	return (&Connector{
		logger:   conRes.Logger(),
	}).Connect
}

// PluginInfo is required as part of the Secretless plugin spec. It provides
// important metadata about the plugin.
func PluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.tcp",
		"id":               "example-tcp-connector",
		"description":      "it's just an example",
	}
}

// GetTCPPlugin is required as part of the Secretless plugin spec for TCP connector
// plugins. It returns the TCP plugin.
func GetTCPPlugin() tcp.Plugin {
	return tcp.ConnectorConstructor(NewConnector)
}
