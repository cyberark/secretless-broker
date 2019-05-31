package v2

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"sort"
)

type ConfigYAML struct {
	Services ServicesYAML
}

type ServiceYAML struct {
	Protocol    string                 `yaml:"protocol" json:"protocol"`
	ListenOn    string                 `yaml:"listenOn" json:"listenOn"`
	Credentials CredentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}            `yaml:"config" json:"config"`
}

// CredentialYAML needs to be an interface because it contains arbitrary YAML
// dictionaries, since end user credential names can be anything.

type CredentialsYAML map[string]interface{}

type ServicesYAML map[string]*ServiceYAML


func NewCredential(credName string, credYAML interface{}) (*Credential, error) {

	cred := &Credential{
		Name: credName,
	}

	// Special Case: string value

	if credVal, ok := credYAML.(string); ok {
		cred.From = "literal"
		cred.Get = credVal
		return cred, nil
	}

	// General Case: provider and id specified

	credentialBytes, err := yaml.Marshal(credYAML)
	if err != nil {
		return nil, err
	}

	credYamlStruct := &struct {
		From string `yaml:"from"`
		Get  string `yaml:"get"`
	}{}

	err = yaml.Unmarshal(credentialBytes, credYamlStruct)
	if err != nil {
		return nil, err
	}

	cred.Get = credYamlStruct.Get
	cred.From = credYamlStruct.From

	return cred, nil
}

func (credentialsYAML *CredentialsYAML) ConvertToCredentials() ([]*Credential, error) {
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


func NewService(svcName string, svcYAML *ServiceYAML) (*Service, error) {

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

func (servicesYAML *ServicesYAML) ToServices() ([]*Service, error) {

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


func NewConfigYAML(fileContents []byte) (*ConfigYAML, error) {
	if len(fileContents) == 0 {
		return nil, fmt.Errorf("empty file contents given to NewConfig")
	}
	cfgYAML := &ConfigYAML{}
	err := yaml.Unmarshal(fileContents, cfgYAML)
	if err != nil {
		return nil, err
	}

	return cfgYAML, nil
}

func (cfgYAML *ConfigYAML) ConvertToConfig() (*Config, error) {
	services, err := cfgYAML.Services.ToServices()
	if err != nil {
		return nil, err
	}

	return &Config{
		Services: services,
	}, nil
}
