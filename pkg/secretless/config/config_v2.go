package config

import (
	"fmt"
	"sort"
	"strings"
)


type Service struct {
	Name             string
	Credentials      []*Credential
	Protocol         string
	ListenOn         string
	Config           []byte
}

type Credential struct {
	Name   string
	From   string
	Get    string
}

type ConfigV2 struct {
	Services []*Service
}

func NewConfigV2(fileContents []byte) (*ConfigV2, error) {
	cfgYAML, err := NewConfigV2YAML(fileContents)
	if err != nil {
		return nil, err
	}

	return cfgYAML.ConvertToConfigV2()
}

func (cfg *ConfigV2) ConvertToV1() (*Config, error) {
	v1Config := &Config{
		Listeners: make([]Listener, 0),
		Handlers:  make([]Handler, 0),
	}

	for _, svc := range cfg.Services  {
		// create listener
		listener := Listener{
			// TODO: find out why this is never used in the Secretless codebase or in example
			// CACertFiles: nil,
			Name:        svc.Name,
			Protocol:    svc.Protocol,
		}

		if strings.HasPrefix(svc.ListenOn, "tcp://") {
			listener.Address = strings.TrimPrefix(svc.ListenOn, "tcp://")
		} else if strings.HasPrefix(svc.ListenOn, "unix://") {
			listener.Socket = strings.TrimPrefix(svc.ListenOn, "unix://")
		} else {
			return nil, fmt.Errorf("convertToV1: listenOn='%s' missing prefix from one of tcp:// or unix//", svc.ListenOn)
		}

		// create handler
		credentials := make([]StoredSecret, 0)
		for _, cred := range svc.Credentials  {
			credentials = append(credentials, StoredSecret{
				Name:     cred.Name,
				Provider: cred.From,
				ID:       cred.Get,
			})
		}
		/// sort Credentials
		sort.Slice(credentials, func(i, j int) bool {
			return credentials[i].Name < credentials[j].Name
		})
		handler := Handler{
			Name:         svc.Name,
			ListenerName: svc.Name,
			Credentials:  credentials,
		}

		// transform listener and handler in a protocol specific way using config
		// TODO: Perhaps this should be injected
		err := transformListenerHandler(svc.Protocol, svc.Config, &listener, &handler)
		if err != nil {
			return nil, err
		}

		// add listener handler pair to v1Config
		v1Config.Listeners = append(v1Config.Listeners, listener)
		v1Config.Handlers = append(v1Config.Handlers, handler)
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

