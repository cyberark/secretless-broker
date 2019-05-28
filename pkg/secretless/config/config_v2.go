package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Service struct {
	Name         string `yaml:"name" json:"name"`
	Protocol     string `yaml:"protocol" json:"protocol"`
	ListenOn string `yaml:"listenOn" json:"listenOn"`
	Credentials  map[string]Credential `yaml:"credentials" json:"credentials"`
	Config       map[string]interface{} `yaml:"config" json:"config"`
}

type Credential struct {
	ProviderId   string `yaml:"providerId" json:"providerId"`
	Provider     string `yaml:"provider" json:"provider"`
}


type ConfigV2 struct {
	Version  string `yaml:"version"`
	Services map[string]Service
}

// TODO: Perhaps rename existing "Config" to "ConfigV1"?
func (cfg *ConfigV2) ConvertToV1() (*Config, error) {
	return nil, nil
}

func NewConfigV2(fileContents []byte) (*ConfigV2, error) {
	if len(fileContents) == 0 {
		return nil, fmt.Errorf("empty file contents given to NewConfig")
	}
	cfg := &ConfigV2{}
	err := yaml.Unmarshal(fileContents, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
