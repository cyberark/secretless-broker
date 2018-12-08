package test

import (
	"fmt"
)

// Represents a simultanous configuration of Secretless and a
// MySQL Server.  This defines a purely logical configuration,
// as opposed to a "live" one which is actually running and
// listening on a port or socket
type AbstractConfiguration struct {
	ListenerType
	ServerTLSType
	SSLModeType
	SSLRootCertType
}

// Represents a "live," running configuration of Secretless and
// MySQL listening on a port or socket.
//
// Rather than creating an "either" type to represent the
// abstract concept of "general portlike thing" -- ie, something
// that can be a TCP port or a Unix socket -- we manually enforce
// the rule that we require exactly 1 of the two.
type LiveConfiguration struct {
	AbstractConfiguration
	port string
	socket string
}

// TODO this should generate ConnectionParams and leave the consumer to implement the flags
// flags is not the first class interface
func (c LiveConfiguration) ConnectionFlags() []string {

	c.validate()

	if c.port != "" {
		return []string{
			fmt.Sprintf("--host=%s", SecretlessHost),
			fmt.Sprintf("--port=%s", c.port),
		}
	} else {
		// sockets take the form /socket/*.sock
		return []string{fmt.Sprintf("--socket=%s", c.socket)}
	}
}

func (c LiveConfiguration) validate() {
	bothEmpty := c.port == "" && c.socket == ""
	bothFilled := c.port != "" && c.socket != ""
	if bothEmpty || bothFilled {
		panic("Corrupted LiveConfiguration, either port or socket but not both")
	}
}


type LiveConfigurations []LiveConfiguration
func (lcs LiveConfigurations) Find(ac AbstractConfiguration) (LiveConfiguration) {
	for _, liveConfiguration := range lcs  {
		if liveConfiguration.AbstractConfiguration == ac {
			return liveConfiguration
		}
	}

	panic(fmt.Errorf("LiveConfiguration not found for AbstractConfiguration: %v", ac))
}
