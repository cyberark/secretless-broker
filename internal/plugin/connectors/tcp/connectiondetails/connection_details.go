package connectiondetails

import (
	"net"
)

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
//
// Any named parameter besides Options is assumed to be required.
type ConnectionDetails struct {
	Host       string
	Port       string
	Username   string
	Password   string
	Options    map[string]string
}

// Handler defines custom functions for handling parameters that are
// not shared by all connectors
type Handler func(credentials map[string][]byte) (map[string]string, error)

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte,
	defaultPort string,
	customHandlers ...Handler) (
	*ConnectionDetails, error) {

	connDetails := &ConnectionDetails{
		Options: make(map[string]string),
	}

	// Required Credentials
	if len(credentials["host"]) > 0 {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = defaultPort
	if len(credentials["port"]) > 0 {
		connDetails.Port = string(credentials["port"])
	}

	if len(credentials["username"]) > 0 {
		connDetails.Username = string(credentials["username"])
	}

	if len(credentials["password"]) > 0 {
		connDetails.Password = string(credentials["password"])
	}

	// Remove required credentials before continuing
	delete(credentials, "host")
	delete(credentials, "port")
	delete(credentials, "username")
	delete(credentials, "password")

	// Add any non-required options to Options
	for key, value := range credentials {
		connDetails.Options[key] = string(value)
	}

	// Perform custom logic using any given handlers
	for _, handler := range customHandlers {
		customOptions, err := handler(credentials)
		if err != nil {
			return nil, err
		}
		for key, value := range customOptions {
			connDetails.Options[key] = value
		}
	}

	return connDetails, nil
}

// Address returns a string representing the network location (host and port)
// of a server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
func (cd *ConnectionDetails) Address() string {
	return net.JoinHostPort(cd.Host, cd.Port)
}
