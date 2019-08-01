package pg

import (
	"log"
	"net"
)

// DefaultPostgresPort is the default port for Postgres database connections
// over TCP
const DefaultPostgresPort = "5432"

var sslOptions = []string{
	"sslrootcert",
	"sslmode",
	"sslkey",
	"sslcert",
}

// BackendConfig stores the connection info to the real backend database.
type BackendConfig struct {
	Host       string
	Port       string
	Username   string
	Password   string
	Options    map[string]string
	SSLOptions map[string]string
}

// Address provides an aggregation of Host and Port fields into a format
// acceptable by consumers of this class (`net.Dial`).
func (backendConfig *BackendConfig) Address() string {
	return net.JoinHostPort(backendConfig.Host, backendConfig.Port)
}

// NewBackendConfig constructs a Backendconfig object based on the options passed
// in that are based on resolved configuration fields.
func NewBackendConfig(options map[string][]byte) (*BackendConfig, error) {
	backendConfig := BackendConfig{
		Options:    make(map[string]string),
		SSLOptions: make(map[string]string),
	}

	if options["host"] != nil {
		backendConfig.Host = string(options["host"])
	}

	backendConfig.Port = DefaultPostgresPort
	if options["port"] != nil {
		backendConfig.Port = string(options["port"])
	}

	// Deprecated. To be removed at a later date and only provided for
	// temporary backwards compatibility.
	if options["address"] != nil {
		log.Printf("WARN: 'address' has been deprecated for PG connector. " +
			"Please use 'host' and 'port' instead.'")

		host, port, err := net.SplitHostPort(string(options["address"]))
		if err != nil {
			// Try one more time but this time assume it's just a hostname
			host, _, err = net.SplitHostPort(string(options["address"]) + ":")
			if err != nil {
				return nil, err
			}
			port = DefaultPostgresPort
		}

		backendConfig.Host = host
		backendConfig.Port = port
	}

	if options["username"] != nil {
		backendConfig.Username = string(options["username"])
	}

	if options["password"] != nil {
		backendConfig.Password = string(options["password"])
	}

	for _, sslOption := range sslOptions {
		if options[sslOption] != nil {
			value := string(options[sslOption])
			if value != "" {
				backendConfig.SSLOptions[sslOption] = value
			}
		}
		delete(options, sslOption)
	}

	delete(options, "host")
	delete(options, "port")
	delete(options, "address")
	delete(options, "username")
	delete(options, "password")

	for k, v := range options {
		backendConfig.Options[k] = string(v)
	}

	return &backendConfig, nil
}
