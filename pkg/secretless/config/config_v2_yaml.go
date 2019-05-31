package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"sort"
)

type ServiceYAML struct {
	Protocol         string                 `yaml:"protocol" json:"protocol"`
	ListenOn         string                 `yaml:"listenOn" json:"listenOn"`
	ProxyCredentials map[string]interface{} `yaml:"credentials" json:"credentials"`
	ProxyConfig      interface{} `yaml:"config" json:"config"`
}

type ConfigV2YAML struct {
	Services map[string]*ServiceYAML
}

type CredentialsYAML struct {
	From   string `yaml:"from"`
	Get     string `yaml:"get"`
}

func NewConfigV2YAML(fileContents []byte) (*ConfigV2YAML, error) {
	if len(fileContents) == 0 {
		return nil, fmt.Errorf("empty file contents given to NewConfig")
	}
	cfgYAML := &ConfigV2YAML{}
	err := yaml.Unmarshal(fileContents, cfgYAML)
	if err != nil {
		return nil, err
	}

	return cfgYAML, nil
}

func (cfgYAML *ConfigV2YAML) ConvertToConfigV2() (*ConfigV2, error)  {
	cfg := &ConfigV2{
		Services: make([]*Service, 0),
	}

	for svcName, svcYAML := range cfgYAML.Services {
		svc := &Service{
			Name:        svcName,
			Credentials: make([]*Credential, 0),
			Protocol:    svcYAML.Protocol,
			ListenOn:    svcYAML.ListenOn,
			Config:      nil,
		}
		for credName, credYAMLAsInterface := range svcYAML.ProxyCredentials {
			cred := &Credential{
				Name: credName,
			}
			if credVal, ok := credYAMLAsInterface.(string); ok {
				cred.From = "literal"
				cred.Get = credVal
			} else {
				credentialBytes, err := yaml.Marshal(credYAMLAsInterface)
				if err != nil {
					return nil, err
				}

				credYAML := &CredentialsYAML{}
				err = yaml.Unmarshal(credentialBytes, credYAML)
				if err != nil {
					return nil, err
				}

				cred.Get = credYAML.Get
				cred.From = credYAML.From

			}
			svc.Credentials = append(svc.Credentials, cred)
		}
		// sort credentials
		sort.Slice(svc.Credentials, func(i, j int) bool {
			return svc.Credentials[i].Name < svc.Credentials[j].Name
		})

		configBytes, err := yaml.Marshal(svcYAML.ProxyConfig)
		if err != nil {
			return nil, err
		}
		svc.Config = configBytes

		cfg.Services = append(cfg.Services, svc)
	}
	// sort services
	sort.Slice(cfg.Services, func(i, j int) bool {
		return cfg.Services[i].Name < cfg.Services[j].Name
	})

	return cfg, nil
}
