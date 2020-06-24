package pg

import (
	"net"
)

// DefaultPostgresPort is the default port for Postgres database connections
// over TCP
const DefaultPostgresPort = "5432"

var sslOptions = []string{
	"host",
	"sslhost",
	"sslrootcert",
	"sslmode",
	"sslkey",
	"sslcert",
}

// ConnectionDetails stores the connection info to the target database.
type ConnectionDetails struct {
	Host       string
	Port       string
	Username   string
	Password   string
	Options    map[string]string
	SSLOptions map[string]string
}

// Address provides an aggregation of Host and Port fields into a format
// acceptable by consumers of this class (`net.Dial`).
func (cd *ConnectionDetails) Address() string {
	return net.JoinHostPort(cd.Host, cd.Port)
}

// NewConnectionDetails constructs a ConnectionDetails object based on the options passed
// in that are based on resolved configuration fields.
func NewConnectionDetails(options map[string][]byte) (*ConnectionDetails, error) {
	connectionDetails := ConnectionDetails{
		Options:    make(map[string]string),
		SSLOptions: make(map[string]string),
	}

	if len(options["host"]) > 0 {
		connectionDetails.Host = string(options["host"])
	}

	connectionDetails.Port = DefaultPostgresPort
	if len(options["port"]) > 0 {
		connectionDetails.Port = string(options["port"])
	}

	// Deprecated. To be removed at a later date and only provided for
	// temporary backwards compatibility.
	if len(options["address"]) > 0 {
		host, port, err := net.SplitHostPort(string(options["address"]))
		if err != nil {
			// Try one more time but this time assume it's just a hostname
			host, _, err = net.SplitHostPort(string(options["address"]) + ":")
			if err != nil {
				return nil, err
			}
			port = DefaultPostgresPort
		}

		connectionDetails.Host = host
		connectionDetails.Port = port
	}

	if len(options["username"]) > 0 {
		connectionDetails.Username = string(options["username"])
	}

	if len(options["password"]) > 0 {
		connectionDetails.Password = string(options["password"])
	}

	for _, sslOption := range sslOptions {
		if len(options[sslOption]) > 0 {
			connectionDetails.SSLOptions[sslOption] = string(options[sslOption])
		}
		delete(options, sslOption)
	}

	delete(options, "host")
	delete(options, "port")
	delete(options, "address")
	delete(options, "username")
	delete(options, "password")

	for k, v := range options {
		connectionDetails.Options[k] = string(v)
	}

	return &connectionDetails, nil
}
