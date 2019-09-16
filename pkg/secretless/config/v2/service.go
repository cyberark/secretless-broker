package v2

import (
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

type serviceYAML struct {
	// Protocol specifies the service connector by protocol.
	// It is an internal detail.
	//
	// Deprecated: Protocol exists for historical compatibility
	// and should not be used. To specify the service connector,
	// use the Connector field.
	Protocol    string          `yaml:"protocol" json:"protocol"`
	Connector   string          `yaml:"connector" json:"connector"`
	ListenOn    string          `yaml:"listenOn" json:"listenOn"`
	Credentials credentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}     `yaml:"config" json:"config"`
}

// Validate verifies the completeness and correctness of the serviceYAML.
func (s serviceYAML) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.ListenOn, validation.Required),
		validation.Field(&s.Credentials, validation.Required),
	)
}

// connectorConfig is a wrapper around byte slice
// that allows the connector configuration
// to be Marshalled to YAML.
type connectorConfig []byte

func (c connectorConfig) MarshalYAML() (interface{}, error) {
	return string(c), nil
}

// Service represents a the configuration of a Secretless proxy service. It
// includes the service's protocol, the socket or address it listens on, the
// location of its required credentials, and (optionally) any additional
// protocol specific configuration.
type Service struct {
	Debug           bool
	Connector       string
	ConnectorConfig connectorConfig
	Credentials     []*Credential
	ListenOn        string
	Name            string
}

// HasCredential indicates whether a Service has the specified credential.
func (s Service) HasCredential(credentialName string) bool {
	for _, credential := range s.Credentials {
		if credential.Name == credentialName {
			return true
		}
	}
	return false
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
	var err error
	errors := validation.Errors{}

	err = svcYAML.Validate()
	if err != nil {
		errors["validation"] = err
	}

	credentials, err := NewCredentials(svcYAML.Credentials)
	if err != nil {
		errors["credentials"] = err
	}

	connectorConfigBytes, err := yaml.Marshal(svcYAML.Config)
	if err != nil {
		errors["config"] = fmt.Errorf("failed to parse 'config' key for service '%s': %s", svcName, err)
	}

	hasConnector := svcYAML.Connector != ""
	hasProtocol := svcYAML.Protocol != ""

	// Protocol given
	if hasProtocol {
		log.Printf("WARN: 'protocol' key found on service '%s'. 'protocol' is now " +
		"deprecated and will be removed in a future release.", svcName)
	}

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
		errors["connector"] = fmt.Errorf("missing 'connector' key on service '%s'", svcName)
	}

	// Accumulate errors from top-level keys on serviceYAML
	err = errors.Filter()
	if err != nil {
		return nil, err
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
