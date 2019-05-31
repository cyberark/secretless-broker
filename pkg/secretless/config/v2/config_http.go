package v2

import (
	"gopkg.in/yaml.v1"
	"github.com/go-ozzo/ozzo-validation"
)

type HTTPConfig struct {
	AuthenticationStrategy   string   `yaml:"authenticationStrategy"`
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

func NewHTTPConfig(cfgBytes []byte) (*HTTPConfig, error) {
	cfg := &HTTPConfig{}
	err := yaml.Unmarshal(cfgBytes, cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *HTTPConfig) Validate() error {
	return validation.ValidateStruct(cfg,
		// AuthenticationStrategy cannot be empty, and must be a recognized strategy
		validation.Field(&cfg.AuthenticationStrategy, validation.Required, validation.In(
			"aws",
			"basic_auth",
			"conjur")),
		// AuthenticateURLsMatching cannot be empty
		validation.Field(&cfg.AuthenticateURLsMatching, validation.Required),
	)
}
