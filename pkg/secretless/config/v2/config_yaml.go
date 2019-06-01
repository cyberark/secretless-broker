package v2

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"sort"
)

type configYAML struct {
	Services servicesYAML
}

type serviceYAML struct {
	Protocol    string          `yaml:"protocol" json:"protocol"`
	ListenOn    string          `yaml:"listenOn" json:"listenOn"`
	Credentials credentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}     `yaml:"config" json:"config"`
}

// CredentialYAML needs to be an interface because it contains arbitrary YAML
// dictionaries, since end user credential names can be anything.

type credentialsYAML map[string]interface{}

type servicesYAML map[string]*serviceYAML

func (credentialsYAML *credentialsYAML) ConvertToCredentials() ([]*Credential, error) {
	credentials := make([]*Credential, 0)

	for credName, credYAML := range *credentialsYAML {
		cred, err := NewCredential(credName, credYAML)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}
	// sort credentials
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	return credentials, nil
}

func NewService(svcName string, svcYAML *serviceYAML) (*Service, error) {

	credentials, err := svcYAML.Credentials.ConvertToCredentials()
	if err != nil {
		return nil, err
	}

	svc := &Service{
		Name:        svcName,
		Credentials: credentials,
		Protocol:    svcYAML.Protocol,
		ListenOn:    svcYAML.ListenOn,
		Config:      nil,
	}

	configBytes, err := yaml.Marshal(svcYAML.Config)
	if err != nil {
		return nil, err
	}
	svc.Config = configBytes

	return svc, nil
}

func (servicesYAML *servicesYAML) ToServices() ([]*Service, error) {

	services := make([]*Service, 0)
	for svcName, svcYAML := range *servicesYAML {
		svc, err := NewService(svcName, svcYAML)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	// sort services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}

func NewConfigYAML(fileContents []byte) (*configYAML, error) {
	if len(fileContents) == 0 {
		return nil, fmt.Errorf("empty file contents given to NewConfig")
	}
	cfgYAML := &configYAML{}
	err := yaml.Unmarshal(fileContents, cfgYAML)
	if err != nil {
		return nil, err
	}

	return cfgYAML, nil
}

func (cfgYAML *configYAML) ConvertToConfig() (*Config, error) {
	services, err := cfgYAML.Services.ToServices()
	if err != nil {
		return nil, err
	}

	return &Config{
		Services: services,
	}, nil
}
