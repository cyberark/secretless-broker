package testutil

import (
	"fmt"
)

// AbstractConfiguration represents a simultaneous configuration of Secretless
// and a Database Server.  This defines a purely logical configuration, as
// opposed to a "live" one which is actually running and listening on a port or
// socket
type AbstractConfiguration struct {
	SocketType
	TLSSetting
	SSLHost
	SSLMode
	RootCertStatus
	PrivateKeyStatus
	PublicCertStatus
	AuthCredentialInvalidity
}

// LiveConfiguration represents a "live," running configuration of Secretless
// and a database listening on a port or socket.
//
// Rather than creating an "either" type to represent the
// abstract concept of "general portlike thing" -- ie, something
// that can be a TCP port or a Unix socket -- we manually enforce
// the rule that we require exactly 1 of the two.
type LiveConfiguration struct {
	AbstractConfiguration
	ConnectionPort
}

// LiveConfigurations is a slice of LiveConfiguration structs.
type LiveConfigurations []LiveConfiguration

// Find returns a LiveConfiguration matching the given AbstractConfiguration.
func (lcs LiveConfigurations) Find(ac AbstractConfiguration) LiveConfiguration {
	for _, liveConfiguration := range lcs {
		if liveConfiguration.AbstractConfiguration == ac {
			return liveConfiguration
		}
	}

	panic(fmt.Errorf("LiveConfiguration not found for AbstractConfiguration: %v", ac))
}

// ConnectionPort represents a TCP port or UNIX socket.
type ConnectionPort struct {
	SocketType
	Port int
}

// Host returns secretless host.
func (cp ConnectionPort) Host() string {
	// NOTE: this is for the good of PG only
	return SecretlessHost
}

// TODO: figure out how to handle the naming convention of unix domain sockets
// for pg. See
// https://www.postgresql.org/docs/9.3/runtime-config-connection.html#GUC-UNIX-SOCKET-DIRECTORIES
// perhaps have test local package pass a NameGenerator that takes the port
// number we'll need to get the port number from the socket file and create the
// appropriate flag

// ToSocketPath returns the path to the pg socket.
func (cp ConnectionPort) ToSocketPath() string {
	// NOTE: this is for the good of PG only
	return fmt.Sprintf("/sock/.s.PGSQL.%v", cp.Port)
}

// ToSocketDir returns the pg socket directory.
func (cp ConnectionPort) ToSocketDir() string {
	return "/sock"
}

// ToPortString returns the port as a string.
func (cp ConnectionPort) ToPortString() string {
	return fmt.Sprintf("%v", cp.Port)
}
