package v2

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"sort"
)

type ServiceYAML struct {
	Protocol    string                 `yaml:"protocol" json:"protocol"`
	ListenOn    string                 `yaml:"listenOn" json:"listenOn"`
	Credentials CredentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}            `yaml:"config" json:"config"`
}

type CredentialsYAML map[string]interface{}
func (credentialsYAML *CredentialsYAML) ConvertToCredentials() ([]*Credential, error) {
	credentials := make([]*Credential, 0)

	for credName, credYAML := range *credentialsYAML {
		cred := &Credential{
			Name: credName,
		}
		if credVal, ok := credYAML.(string); ok {
			cred.From = "literal"
			cred.Get = credVal
		} else {
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

		}
		credentials = append(credentials, cred)
	}
	// sort credentials
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	return credentials, nil
}

type ServicesYAML map[string]*ServiceYAML

func (servicesYAML *ServicesYAML) ConvertToServices() ([]*Service, error) {
	services := make([]*Service, 0)
	for svcName, svcYAML := range *servicesYAML {
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

		services = append(services, svc)
	}
	// sort services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}

type ConfigYAML struct {
	Services ServicesYAML
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
	services, err := cfgYAML.Services.ConvertToServices()
	if err != nil {
		return nil, err
	}

	return &Config{
		Services: services,
	}, nil
}
