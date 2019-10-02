package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

type examplePlugin struct{}

// connectorFunc is the function that implements the tcp.Connector func signature in this
// example plugin. It has access to the client connection and the credentials (as a map),
// and is expected to return the target service connection.
//
// This example connector works as follows:
// 1. Waits for the initial message from the client
// 2. Connect to a target service whose address is the value of the credential identified
// by the key "address"
// 3. Inject credentials from a credential identified by the key "auth"
// 4. Write the initial message from the client with some modification
func connectorFunc(clientConn net.Conn, credentialValuesByID connector.CredentialValuesByID) (net.Conn, error) {
	clientInitMsg, _, err := bufio.NewReader(clientConn).ReadLine()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", string(credentialValuesByID["address"]))
	if err != nil {
		return nil, err
	}

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

	return conn, nil
}

// NewConnector is required method on the tcp.Plugin interface. It returns a
// tcp.Connector.
//
// The single argument passed in is of type connector.Resources. It contains
// connector-specific config and a logger. This particular plugin ignores these.
func (examplePlugin *examplePlugin) NewConnector(_ connector.Resources) tcp.Connector {
	return connectorFunc
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
	return &examplePlugin{}
}
