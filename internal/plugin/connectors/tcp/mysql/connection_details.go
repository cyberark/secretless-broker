package mysql

import "strconv"

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
type ConnectionDetails struct {
	Host     string
	Options  map[string]string
	Password string
	Port     uint
	Username string
}

const DefaultMySQLPort = uint(3306)

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte) (
	*ConnectionDetails, error) {

	connDetails := &ConnectionDetails{Options: make(map[string]string)}

	if host := credentials["host"]; host != nil {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = DefaultMySQLPort
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

	// Make sure that we process the SSL mode arg even if it's not specified
	// otherwise it will get ignored
	if _, ok := credentials["sslmode"]; !ok {
		credentials["sslmode"] = []byte("")
	}

	delete(credentials, "host")
	delete(credentials, "port")
	delete(credentials, "username")
	delete(credentials, "password")

	connDetails.Options = make(map[string]string)
	for k, v := range credentials {
		connDetails.Options[k] = string(v)
	}

	return connDetails, nil
}

// Address returns a string representing the network location (host and port)
// of a MySQL server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
func (cd *ConnectionDetails) Address() string {
	return cd.Host + ":" + strconv.FormatUint(uint64(cd.Port), 10)
}
