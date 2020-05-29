package cassandra

import "strconv"

// DefaultCqlPort is the default port for Cql database connections
// over TCP
const DefaultCqlPort = 9042

var sslOptions = []string{
	"sslrootcert",
	"sslmode",
	"sslkey",
	"sslcert",
}

// ConnectionDetails stores the connection info to the target database.
type ConnectionDetails struct {
	Host       string
	Port       int
	Username   string
	Password   string
	sslOptions map[string]string
}

const (
	sslModeDisable    = "disable"
	sslModeRequire    = "require"
	sslModeVerifyCA   = "verify-ca"
	sslModeVerifyFull = "verify-full"
)

// NewConnectionDetails constructs a ConnectionDetails object based on the options passed
// in that are based on resolved configuration fields.
func NewConnectionDetails(options map[string][]byte) (*ConnectionDetails, error) {
	connectionDetails := ConnectionDetails{
		sslOptions: map[string]string{},
	}

	if len(options["host"]) > 0 {
		connectionDetails.Host = string(options["host"])
	}

	connectionDetails.Port = DefaultCqlPort
	if len(options["port"]) > 0 {
		portStr := string(options["port"])
		connectionDetails.Port, _ = strconv.Atoi(portStr)
	}

	if len(options["username"]) > 0 {
		connectionDetails.Username = string(options["username"])
	}

	if len(options["password"]) > 0 {
		connectionDetails.Password = string(options["password"])
	}

	for _, option := range sslOptions {
		if value := options[option]; len(value) > 0 {
			connectionDetails.sslOptions[option] = string(value)
		}
	}

	return &connectionDetails, nil
}
