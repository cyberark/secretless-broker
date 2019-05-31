package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Service struct {
	Name             string
	Protocol         string
	ListenOn         string
	Credentials      []Credential
	Config           []byte
}

type Credential struct {
	Name   string `yaml:"-"`
	From   string `yaml:"from" json:"from"`
	Get    string `yaml:"get" json:"get"`
}

type HttpConfig struct {
	authenticationStrategy string
	authenticateURLsMatching []string
}


type ConfigV2 struct {
	Version  string `yaml:"version"`
	Services []Service
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
