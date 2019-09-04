package v2

import (
	"fmt"
	"log"
	"sort"

	validation "github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

// Config represents a full configuration of Secretless, which is just a list of
// individual Service configurations.
type Config struct {
	Debug bool
	Services []*Service
}

// Validate verifies the completeness and correctness of the Config.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Services, validation.Required),
	)
}

// Serialize Config to YAML
func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewConfig creates a v2.Config from yaml bytes
func NewConfig(v2YAML []byte) (*Config, error) {
	cfgYAML, err := newConfigYAML(v2YAML)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0)
	for svcName, svcYAML := range cfgYAML.Services {
		svc, err := NewService(svcName, svcYAML)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	// sort Services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return &Config{
		Services: services,
	}, nil
}

// connectorFromLegacyHTTPConfig extracts authenticationStrategy.
// This function is useful when the deprecated 'protocol' field equals 'http'
// and you want to determine the connector name
func connectorFromLegacyHTTPConfig(connectorConfigBytes []byte) (string, error) {
	tempCfg := &struct {
		AuthenticationStrategy string `yaml:"authenticationStrategy"`
	}{}
	err := yaml.Unmarshal(connectorConfigBytes, tempCfg)
	if err != nil {
		return "", err
	}

	err = validation.ValidateStruct(
		tempCfg,
		validation.Field(
			&tempCfg.AuthenticationStrategy,
			validation.Required,
			validation.In(HTTPAuthenticationStrategies...),
		),
	)

	if err != nil {
		return "", err
	}

	return tempCfg.AuthenticationStrategy, nil
}

// NewService creates a named v2.Service from yaml bytes
func NewService(svcName string, svcYAML *serviceYAML) (*Service, error) {
	credentials, err := NewCredentials(svcYAML.Credentials)
	if err != nil {
		return nil, err
	}

	connectorConfigBytes, err := yaml.Marshal(svcYAML.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'config' key for service '%s': %s", svcName, err)
	}

	hasConnector := svcYAML.Connector != ""
	hasProtocol := svcYAML.Protocol != ""

	// Both connector and protocol given
	if hasConnector && hasProtocol {
		log.Printf("WARN: 'connector' and 'protocol' keys found on "+
			"service '%s'. 'connector' key takes precendence, 'protocol' is "+
			"deprecated.", svcName)
	}

	var connector string

	// Connector given, always takes precedence
	if hasConnector {
		connector = svcYAML.Connector

	// Only use protocol when connector not given
	} else if hasProtocol {
		connector = svcYAML.Protocol

	// Neither given
	} else {
		return nil, fmt.Errorf("missing 'connector' key on service '%s'", svcName)
	}

	// When only the deprecated 'protocol' field
	// is given and it equals 'http' the connector name
	// must be extracted from the http config
	if !hasConnector && hasProtocol && connector == "http" {
		connector, err = connectorFromLegacyHTTPConfig(connectorConfigBytes)
		if err != nil {
			return nil, fmt.Errorf("error on http config for service '%s': %s", svcName, err)
		}
	}

	return &Service{
		Credentials:     credentials,
		ListenOn:        svcYAML.ListenOn,
		Name:            svcName,
		Connector:       connector,
		ConnectorConfig: connectorConfigBytes,
	}, nil
}
