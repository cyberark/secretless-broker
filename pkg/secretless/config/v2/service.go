package v2

import (
	"fmt"
	"log"
	"net/url"
	"os"
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

// ConfigEnv represents the runtime environment that will fulfill the services
// requested by the Config.  It has a single public method, Prepare, that
// ensures the runtime environment supports the requested services.
type ConfigEnv struct {
	logger logapi.Logger
	availPlugins plugin.AvailablePlugins
	getFileInfo func(name string) (os.FileInfo, error)
	deleteFile func(name string) error
}

// NewConfigEnv creates a new instance of ConfigEnv.
func NewConfigEnv(logger logapi.Logger, availPlugins plugin.AvailablePlugins) ConfigEnv {
	// Return *PathError
	// &PathError{"remove", name, e}
	// os.Remove("blah")
	return NewConfigEnvWithOptions(logger, availPlugins, os.Stat, os.Remove)
}

// NewConfigEnvWithOptions allows injecting all dependencies.  Used for unit
// testing.
func NewConfigEnvWithOptions(
	logger logapi.Logger,
	availPlugins plugin.AvailablePlugins,
	getFileInfo func(name string) (os.FileInfo, error),
	deleteFile func(name string) error,
) ConfigEnv {
	// Return *PathError
	// &PathError{"remove", name, e}
	os.Remove("blah")
	return ConfigEnv{
		logger:       logger,
		availPlugins: availPlugins,
		getFileInfo: getFileInfo,
		deleteFile: deleteFile,
	}
}
// This is just a pure type error, we can create our own using os.ErrNotExist
//func IsNotExist(err error) bool {
//func Stat(name string) (FileInfo, error) {
//func Remove(name string) error {
//type FileInfo interface {
//	Name() string       // base name of the file
//	Size() int64        // length in bytes for regular files; system-dependent for others
//	Mode() FileMode     // file mode bits
//	ModTime() time.Time // modification time
//	IsDir() bool        // abbreviation for Mode().IsDir()
//	Sys() interface{}   // underlying data source (can return nil)
//}

// Prepare ensures the runtime environment is prepared to handle the Config's
// service requests. It checks both that the requested connectors exist, and
// that the requested sockets are available, or can be deleted.  If any of these
// checks fail, it will error.
func (c *ConfigEnv) Prepare(cfg Config) error {
	err := c.validateRequestedPlugins(cfg)
	if err != nil {
		return err
	}

	return c.ensureAllSocketsAreDeleted(cfg)
}

// validateRequestedPlugins ensures that the AvailablePlugins can fulfill the
// services requested by the given Config, and return an error if not.
func (c *ConfigEnv) validateRequestedPlugins(cfg Config) error {
	pluginIDs := plugin.AvailableConnectorIDs(c.availPlugins)

	c.logger.Infof(
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

func (c *ConfigEnv) ensureAllSocketsAreDeleted(cfg Config) error {
	errors := validation.Errors{}

	for _, service := range cfg.Services {
		err := c.ensureSocketIsDeleted(service.ListenOn)
		if err != nil {
			errors[service.Name] = fmt.Errorf(
				"socket can't be deleted: %s", service.ListenOn,
			)
		}
	}
	return errors.Filter()
}

func (c *ConfigEnv) ensureSocketIsDeleted(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("unable to parse ListenOn location '%s'", address)
	}

	// If we're not a unix socket address, we don't need to worry about pre-emptive cleanup
	if parsedURL.Scheme != "unix" {
		return nil
	}

	socketFile := parsedURL.Path
	c.logger.Debugf("Ensuring that the socketfile '%s' is not present...", socketFile)

	// If file is not present, then we are ok to continue.
	// NOTE: os.IsNotExist is a pure function, so does not need to be injected.
	if _, err := c.getFileInfo(socketFile); os.IsNotExist(err) {
		c.logger.Debugf("Socket file '%s' not present. Skipping deletion.", socketFile)
		return nil
	}

	// Otherwise delete the file first
	c.logger.Warnf("Socket file '%s' already present. Deleting...", socketFile)
	err = c.deleteFile(socketFile)
	if err != nil {
		return fmt.Errorf("unable to delete stale socket file '%s'", socketFile)
	}

	return nil
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
