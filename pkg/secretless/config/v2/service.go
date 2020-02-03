package v2

import (
	"fmt"
	"log"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type serviceYAML struct {
	// Protocol specifies the service connector by protocol. It is an internal
	// detail.
	//
	// Deprecated: Protocol exists for historical compatibility and should not
	// be used. To specify the service connector, use the Connector field.
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

// connectorConfig is a wrapper around byte slice that allows the connector
// configuration to be Marshalled to YAML.
type connectorConfig []byte

func (c connectorConfig) MarshalYAML() (interface{}, error) {
	var out interface{}
	err := yaml.Unmarshal(c, &out)
	if err != nil {
		err = errors.Wrap(err, "failed to marshal connectorConfig to YAML")
	}
	return out, err
}

// Service represents the configuration of a Secretless proxy service. It
// includes the service's protocol, the socket or address it listens on, the
// location of its required credentials, and (optionally) any additional
// protocol specific configuration.
type Service struct {
	Debug           bool
	Connector       string
	ConnectorConfig connectorConfig
	Credentials     []*Credential
	ListenOn        NetworkAddress
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

// connectorFromLegacyHTTPConfig extracts authenticationStrategy. This function
// is useful when the deprecated 'protocol' field equals 'http' and you want to
// determine the connector name
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
		errors["config"] = fmt.Errorf(
			"failed to parse 'config' key for service '%s': %s",
			svcName,
			err,
		)
	}

	connector, err := connectorID(svcName, svcYAML, connectorConfigBytes)
	if err != nil {
		errors["connector"] = err
	}

	// Accumulate errors from top-level keys on serviceYAML
	err = errors.Filter()
	if err != nil {
		return nil, err
	}

	return &Service{
		Credentials:     credentials,
		ListenOn:        NetworkAddress(svcYAML.ListenOn),
		Name:            svcName,
		Connector:       connector,
		ConnectorConfig: connectorConfigBytes,
	}, nil
}

// connectorID determines the connectorID or errors.  It handles identification
// of the deprecated 'protocol' field, and issues appropriate warnings.
func connectorID(
	svcName string,
	svcYAML *serviceYAML,
	connectorConfigBytes []byte,
) (string, error) {
	var err error
	hasConnector := svcYAML.Connector != ""
	hasProtocol := svcYAML.Protocol != ""

	// Protocol given
	if hasProtocol {
		log.Printf(
			"WARN: 'protocol' key found on service '%s'. 'protocol' is now "+
				"deprecated and will be removed in a future release.",
			svcName,
		)
	}

	// Both connector and protocol given
	if hasConnector && hasProtocol {
		log.Printf("WARN: 'connector' and 'protocol' keys found on "+
			"service '%s'. 'connector' key takes precendence, 'protocol' is "+
			"deprecated.", svcName)
	}

	var connectorID string
	switch {
	case hasConnector: // Connector given, always takes precedence
		connectorID = svcYAML.Connector
	case hasProtocol: // Only use protocol when connector not given
		connectorID = svcYAML.Protocol
	default:
		return "", fmt.Errorf("missing 'connector' key on service '%s'", svcName)
	}

	// When only the deprecated 'protocol' field is given and it equals 'http'
	// the connector name must be extracted from the http config
	if !hasConnector && hasProtocol && connectorID == "http" {
		connectorID, err = connectorFromLegacyHTTPConfig(connectorConfigBytes)
		if err != nil {
			return "", fmt.Errorf(
				"error on http config for service '%s': %s",
				svcName,
				err,
			)
		}
	}

	return connectorID, nil
}

// NetworkAddress is a utility type for handling string manipulation /
// destructuring for listenOn addresses that include a network. Currently only
// used outside this package.
type NetworkAddress string

// Network returns the "network" part of a network address, eg, "tcp" or "unix".
func (a NetworkAddress) Network() string {
	return a.split()[0]
}

// Address returns the "address" part of a network address, eg, "127.0.0.1".
func (a NetworkAddress) Address() string {
	return a.split()[1]
}

func (a NetworkAddress) split() []string {
	return strings.Split(string(a), "://")
}
