package v2

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

// newV2Service translates an associated v1 Listener-Handler pair to a v2 Service.
// This method illustrates how the conceptual model of V2 Services combines
// the legacy concept of Handlers and Listeners into a singular entity.
func newV2Service(listener config_v1.Listener, linkedHandler config_v1.Handler) (*Service, error) {
	// Extract Connector and ConnectorConfig
	var connectorConfig []byte

	connector := listener.Protocol
	if connector == "http" {
		var err error
		connector = strings.TrimPrefix(linkedHandler.Type, "http/")

		tempCfg := &struct {
			AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
		}{
			AuthenticateURLsMatching: linkedHandler.Match,
		}

		connectorConfig, err = yaml.Marshal(tempCfg)
		if err != nil {
			return nil, err
		}
	}

	// Extract ListenOn
	listenOn := fmt.Sprintf("tcp://%s", listener.Address)
	if listener.Address == "" {
		listenOn = fmt.Sprintf("unix://%s", listener.Socket)
	}

	// Extract Credentials
	credentials := make([]*Credential, 0)
	for _, storedSecret := range linkedHandler.Credentials {
		credentials = append(credentials, &Credential{
			Name: storedSecret.Name,
			From: storedSecret.Provider,
			Get:  storedSecret.ID,
		})
	}
	// Sort Credentials
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	// Create Service
	return &Service{
		Connector:       connector,
		ConnectorConfig: connectorConfig,
		Credentials:     credentials,
		ListenOn:        listenOn,
		Name:            linkedHandler.Name,
	}, nil
}

func newV2Config(v1Cfg *config_v1.Config) (*Config, error) {
	// Validate v1 Config
	if err := v1Cfg.Validate(); err != nil {
		return nil, err
	}

	// Create list of v2 Services
	v2Services := make([]*Service, 0)
	for _, listener := range v1Cfg.Listeners {
		linkedHandlers := listener.LinkedHandlers(v1Cfg.Handlers)

		// Non-http listeners only use the first linked handler
		if listener.Protocol != "http" {
			linkedHandlers = linkedHandlers[0:1]
		}

		// Create v2 Service from each v1 Listener-Handler pair
		for _, linkedHandler := range linkedHandlers {
			v2Service, err := newV2Service(listener, linkedHandler)
			if err != nil {
				return nil, err
			}

			v2Services = append(v2Services, v2Service)
		}
	}

	// Sort Services on Name
	sort.Slice(v2Services, func(i, j int) bool {
		return v2Services[i].Name < v2Services[j].Name
	})

	return &Config{
		Services: v2Services,
	}, nil
}
