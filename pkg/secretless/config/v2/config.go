package v2

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	"sort"
	"strings"
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

func (svc *Service) ToV1Service() (*v1Service, error) {

	// Create listener

	listener := config.Listener{
		// TODO: find out why this is never used in the Secretless codebase or in example
		//  CACertFiles: nil
		Name:     svc.Name,
		Protocol: svc.Protocol,
	}

	// Convert listenOn to Address or Socket, depending on protocol

	if strings.HasPrefix(svc.ListenOn, "tcp://") {
		listener.Address = strings.TrimPrefix(svc.ListenOn, "tcp://")
	} else if strings.HasPrefix(svc.ListenOn, "unix://") {
		listener.Socket = strings.TrimPrefix(svc.ListenOn, "unix://")
	} else {
		errMsg := "listenOn=%q missing prefix from one of tcp:// or unix//"
		return nil, fmt.Errorf(errMsg, svc.ListenOn)
	}

	// Create handler

	credentials := make([]config.StoredSecret, 0)
	for _, cred := range svc.Credentials {
		credentials = append(credentials, config.StoredSecret{
			Name:     cred.Name,
			Provider: cred.From,
			ID:       cred.Get,
		})
	}

	// Sort Credentials

	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})
	handler := config.Handler{
		Name:         svc.Name,
		ListenerName: svc.Name,
		Credentials:  credentials,
	}

	// Apply protocol specific config

	v1Service := &v1Service{&listener, &handler}
	err := v1Service.applyProtocolConfig(svc.Config)
	if err != nil {
		return nil, err
	}

	return v1Service, nil
}

type Credential struct {
	Name string
	From string
	Get  string
}

func NewConfig(fileContents []byte) (*Config, error) {
	cfgYAML, err := NewConfigYAML(fileContents)
	if err != nil {
		return nil, err
	}

	return cfgYAML.ConvertToConfig()
}

func (cfg *Config) ConvertToV1() (*config.Config, error) {
	v1Config := &config.Config{
		Listeners: make([]config.Listener, 0),
		Handlers:  make([]config.Handler, 0),
	}

	for _, svc := range cfg.Services {
		v1Svc, err := svc.ToV1Service()
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
