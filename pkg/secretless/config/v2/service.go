package v2

import (
	"fmt"
	"log"
	"strings"

	logapi "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	validation "github.com/go-ozzo/ozzo-validation"
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
	return string(c), nil
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
		ListenOn:        svcYAML.ListenOn,
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
			"WARN: 'protocol' key found on service '%s'. 'protocol' is now " +
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
	case hasProtocol:  // Only use protocol when connector not given
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

// ConfigEnv represents the runtime environment that will fulfill the requested
// connectors. Its static dependencies are a logger and the plugins available to
// create connectors.
// TODO: We'll move the logic from ensureSocketIsDeleted into here soon.
type ConfigEnv struct {
	logger logapi.Logger
	availPlugins plugin.AvailablePlugins
}

// NewConfigEnv creates a new instance of ConfigEnv.
func NewConfigEnv(l logapi.Logger, a plugin.AvailablePlugins) ConfigEnv {
	return ConfigEnv{
		logger:       l,
		availPlugins: a,
	}
}

// Prepare ensures the runtime environment is prepared to handle the Config's
// service requests. Currently, this just means it checks that the requested
// connectors exist, based on the AvailablePlugins.
func (v *ConfigEnv) Prepare(cfg Config) error {
	pluginIDs := plugin.AvailableConnectorIDs(v.availPlugins)

	v.logger.Infof(
		"Validating config against available plugins: %s",
		strings.Join(pluginIDs, ","),
	)

	// Convert available plugin IDs to a map, so that we can check if they exist
	// in the loop below using a map lookup rather than a nested loop.
	pluginIDsMap := map[string]bool{}
	for _, p := range pluginIDs {
		pluginIDsMap[p] = true
	}

	errors := validation.Errors{}
	for _, service := range cfg.Services {
		// A plugin ID and a connector name are equivalent.
		pluginExists := pluginIDsMap[service.Connector]
		if !pluginExists {
			errors[service.Name] = fmt.Errorf(
				`missing service connector "%s"`,
				service.Connector,
			)
			continue
		}
	}

	err := errors.Filter()
	if err != nil {
		err = fmt.Errorf("services validation failed: %s", err.Error())
	}

	return err
}

// NetworkAddress is a utility type for handling string manipulation /
// destructuring for listenOn addresses that include a network. Currently only
// used outside this package.
// TODO: Update all instances of listenOn to use this type
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
