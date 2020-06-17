package mysql

import "strconv"

// DefaultMySQLPort is the default port on which we connect to the MySQL service
// If another port is found within the connectionDetails, we will use that.
const DefaultMySQLPort = uint(3306)

var sslOptions = []string{
	"host",
	"sslhost",
	"sslrootcert",
	"sslmode",
	"sslkey",
	"sslcert",
}

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
type ConnectionDetails struct {
	Host       string
	Options    map[string]string
	Password   string
	Port       uint
	SSLOptions map[string]string
	Username   string
}

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte) (
	*ConnectionDetails, error) {

	connDetails := &ConnectionDetails{
		Options:    make(map[string]string),
		SSLOptions: make(map[string]string),
	}

	if len(credentials["host"]) > 0 {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = DefaultMySQLPort
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

	for _, sslOption := range sslOptions {
		if len(credentials[sslOption]) > 0 {
			connDetails.SSLOptions[sslOption] = string(credentials[sslOption])
		}
		delete(credentials, sslOption)
	}

	delete(credentials, "host")
	delete(credentials, "port")
	delete(credentials, "username")
	delete(credentials, "password")

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
