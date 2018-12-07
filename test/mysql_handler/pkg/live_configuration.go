package pkg

import (
	"fmt"
)

type AbstractConfiguration struct {
	ListenerType
	ServerTLSType
	SSLModeType
	SSLRootCertType
}

type LiveConfiguration struct {
	AbstractConfiguration
	port string
	socket string
}

func (c LiveConfiguration) ConnectionFlags() []string {
	if c.port != "" && c.socket != "" {
		panic("Corrupted LiveConfiguration, either port or socket not both")
	}

	if c.port != "" {
		return []string{
			fmt.Sprintf("--host=%s", SecretlessHost),
			fmt.Sprintf("--port=%s", c.port),
		}
	}
	if c.socket != "" {
		// sockets take the form /socket/*.sock
		return []string{fmt.Sprintf("--socket=%s", c.socket)}
	}

	panic("Corrupted LiveConfiguration, either port or socket not none")
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
