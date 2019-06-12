package v2

import (
	"sort"

	"gopkg.in/yaml.v2"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

// Config represents a full configuration of Secretless, which is just a list of
// individual Service configurations.
type Config struct {
	Services []*Service
}

// Service represents a the configuration of a Secretless proxy service. It
// includes the service's protocol, the socket or address it listens on, the
// location of its required credentials, and (optionally) any additional
// protocol specific configuration.
type Service struct {
	Credentials    []*Credential
	ListenOn       string
	Name           string
	Protocol       string
	ProtocolConfig []byte
}

// NewV1Config converts the bytes of a v2 YAML file to a v1.Config.  As such,
// it's the primary public interface of the v2 package, and probably only
// func most users will need.
func NewV1Config(v2YAML []byte) (*config_v1.Config, error) {
	v2cfg, err := NewConfig(v2YAML)
	if err != nil {
		return nil, err
	}

	v1cfg, err := NewV1ConfigFromV2Config(v2cfg)
	if err != nil {
		return nil, err
	}

	return v1cfg, nil
}

// NewV1ConfigFromV2Config converts a v2.Config to a v1.Config.
func NewV1ConfigFromV2Config(v2cfg *Config) (*config_v1.Config, error) {
	v1Config := &config_v1.Config{
		Listeners: make([]config_v1.Listener, 0),
		Handlers:  make([]config_v1.Handler, 0),
	}

	for _, svc := range v2cfg.Services {
		v1Svc, err := newV1Service(*svc)
		if err != nil {
			return nil, err
		}
		v1Config.Listeners = append(v1Config.Listeners, *v1Svc.Listener)
		v1Config.Handlers = append(v1Config.Handlers, *v1Svc.Handler)
	}

	// sort Listeners
	sort.Slice(v1Config.Listeners, func(i, j int) bool {
		return v1Config.Listeners[i].Name < v1Config.Listeners[j].Name
	})

	// sort Handlers
	sort.Slice(v1Config.Handlers, func(i, j int) bool {
		return v1Config.Handlers[i].Name < v1Config.Handlers[j].Name
	})

	return v1Config, nil
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

	return &Config{
		Services: services,
	}, nil
}

// NewService creates a named v2.Service from yaml bytes
func NewService(svcName string, svcYAML *serviceYAML) (*Service, error) {
	credentials, err := NewCredentials(svcYAML.Credentials)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		Credentials:    credentials,
		ListenOn:       svcYAML.ListenOn,
		Name:           svcName,
		Protocol:       svcYAML.Protocol,
		ProtocolConfig: nil,
	}

	configBytes, err := yaml.Marshal(svcYAML.Config)
	if err != nil {
		return nil, err
	}
	svc.ProtocolConfig = configBytes

	return svc, nil
}
