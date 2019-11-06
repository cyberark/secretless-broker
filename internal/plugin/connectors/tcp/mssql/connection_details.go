package mssql

import (
	"strconv"
)

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
type ConnectionDetails struct {
	Host     string
	Port     uint
	Username string
	Password string
}

const defaultMSSQLPort = uint(1433)

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte) (*ConnectionDetails, error) {

	connDetails := &ConnectionDetails{}

	if host := credentials["host"]; host != nil {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = defaultMSSQLPort
	if credentials["port"] != nil {
		port64, _ := strconv.ParseUint(string(credentials["port"]), 10, 64)
		connDetails.Port = uint(port64)
	}

	if credentials["username"] != nil {
		connDetails.Username = string(credentials["username"])
	}

	if credentials["password"] != nil {
		connDetails.Password = string(credentials["password"])
	}

	return connDetails, nil
}

// Address returns a string representing the network location (host and port)
// of a MSSQL server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
func (cd *ConnectionDetails) Address() string {
	return cd.Host + ":" + strconv.FormatUint(uint64(cd.Port), 10)
}
