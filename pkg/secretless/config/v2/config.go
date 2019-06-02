// TODO: find out why this is never used in the Secretless codebase or in example
//  CACertFiles: nil

package v2

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"gopkg.in/yaml.v2"
	"sort"
)

type Config struct {
	Services []*Service
}

type Service struct {
	Name           string
	Credentials    []*Credential
	Protocol       string
	ListenOn       string
	ProtocolConfig []byte
}

// NewV1Config is converts the bytes of a v2 YAML file to a v1.Config.  As such,
// it's the primary public interface of the v2 package.
// TODO: Possible move to v1.config?
func NewV1Config(v2YAML []byte) (*v1.Config, error) {
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

func NewV1ConfigFromV2Config(v2cfg *Config) (*v1.Config, error) {
	v1Config := &v1.Config{
		Listeners: make([]v1.Listener, 0),
		Handlers:  make([]v1.Handler, 0),
	}

	for _, svc := range v2cfg.Services {
		v1Svc, err := NewV1Service(*svc)
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

func NewService(svcName string, svcYAML *serviceYAML) (*Service, error) {

	credentials, err := NewCredentials(svcYAML.Credentials)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		Name:           svcName,
		Credentials:    credentials,
		Protocol:       svcYAML.Protocol,
		ListenOn:       svcYAML.ListenOn,
		ProtocolConfig: nil,
	}

	configBytes, err := yaml.Marshal(svcYAML.Config)
	if err != nil {
		return nil, err
	}
	svc.ProtocolConfig = configBytes

	return svc, nil
}
