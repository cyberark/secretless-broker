// TODO: find out why this is never used in the Secretless codebase or in example
//  CACertFiles: nil

package v2

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"sort"
)

type Config struct {
	Services []*Service
}

type Service struct {
	Name        string
	Credentials []*Credential
	Protocol    string
	ListenOn    string
	Config      []byte
}

type Credential struct {
	Name string
	From string
	Get  string
}

func NewV1Config(fileContents []byte) (*v1.Config, error) {
	v2cfg, err := NewConfig(fileContents)
	if err != nil {
		return nil, err
	}

	v1cfg, err := v2cfg.ConvertToV1()
	if err != nil {
		return nil, err
	}

	return v1cfg, nil
}

func NewConfig(fileContents []byte) (*Config, error) {
	cfgYAML, err := NewConfigYAML(fileContents)
	if err != nil {
		return nil, err
	}

	return cfgYAML.ConvertToConfig()
}

func (cfg *Config) ConvertToV1() (*v1.Config, error) {
	v1Config := &v1.Config{
		Listeners: make([]v1.Listener, 0),
		Handlers:  make([]v1.Handler, 0),
	}

	for _, svc := range cfg.Services {
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
