package main

// TODO: change the package name to your plugin name if this will be an internal connector

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// SingleUseConnector is passed the client's net.Conn and the current CredentialValuesById,
// and returns an authenticated net.Conn to the target service
type SingleUseConnector struct {
	logger log.Logger
	config []byte // Note: this can be removed if your plugin does not use any custom config
}

// Connect receives a connection to the client, and opens a connection to the target using the client's connection
// and the credentials provided in credentialValuesByID
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	// TODO: add logic according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#tcp-connector
	// tcp/pg/connector.go is a good example.

	var err error
	return nil, err
}
