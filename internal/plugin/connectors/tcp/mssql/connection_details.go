package mssql

import (
	"strconv"
)

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
type ConnectionDetails struct {
	Host       string
	Port       uint
	Username   string
	Password   string
	SSLMode    string
	SSLOptions map[string]string
}

var sslOptions = []string{
	"sslrootcert",
	"sslkey",
	"sslcert",
}

const defaultMSSQLPort = uint(1433)

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte) *ConnectionDetails {

	connDetails := &ConnectionDetails{}

	if len(credentials["host"]) > 0 {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = defaultMSSQLPort
	if len(credentials["port"]) > 0 {
		port64, _ := strconv.ParseUint(string(credentials["port"]), 10, 64)
		connDetails.Port = uint(port64)
	}

	if len(credentials["username"]) > 0 {
		connDetails.Username = string(credentials["username"])
	}

	if len(credentials["password"]) > 0 {
		connDetails.Password = string(credentials["password"])
	}

	sslMode := string(credentials["sslmode"])
	if sslMode != "disable" {
		// Currently, we only support disable
		sslMode = "disable"
		// In the event that the user does not choose 'disable', i.e. when
		// ssl is 'required', we want to parse the additional ssl credentials
		// that are are needed
		connDetails.SSLOptions = newSSLOptions(credentials)
	}

	connDetails.SSLMode = sslMode

	return connDetails
}

// Address returns a string representing the network location (host and port)
// of a MSSQL server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
func (cd *ConnectionDetails) Address() string {
	return cd.Host + ":" + strconv.FormatUint(uint64(cd.Port), 10)
}

func newSSLOptions(credentials map[string][]byte) map[string]string {
	SSLOptions := make(map[string]string)

	for _, sslOption := range sslOptions {
		value := string(credentials[sslOption])
		if len(value) > 0 {
			SSLOptions[sslOption] = value
		}
	}

	return SSLOptions
}
