package v2

import (
	"gopkg.in/yaml.v1"
)

type HTTPConfig struct {
	AuthenticationStrategy   string   `yaml:"authenticationStrategy"`
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

func NewHTTPConfig(cfgBytes []byte) (HTTPConfig, error) {
	cfg := &HTTPConfig{}
	err := yaml.Unmarshal(cfgBytes, cfg)

	return *cfg, err
}
