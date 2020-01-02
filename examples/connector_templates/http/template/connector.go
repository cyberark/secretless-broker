package main

// TODO: change the package name to your plugin name if this will be an internal connector

import (
	gohttp "net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Connector modifies an HTTP request to include required authentication information
type Connector struct {
	logger log.Logger
	config []byte // Note: this can be removed if your plugin does not use any custom config
}

/*
Connect has access to the client http.Request and the credentials
(as a map), and is expected to modify the request so that it will authenticate.
This typically means adding required authorization headers.
*/
func (c *Connector) Connect(
	r *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	// TODO: add logic according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#http-connector

	var err error
	return err
}
