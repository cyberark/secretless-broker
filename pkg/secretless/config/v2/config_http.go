package v2

import (
	"gopkg.in/yaml.v1"
	"github.com/go-ozzo/ozzo-validation"
)

type httpConfig struct {
	AuthenticationStrategy   string   `yaml:"authenticationStrategy"`
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

func newHTTPConfig(cfgBytes []byte) (*httpConfig, error) {
	cfg := &httpConfig{}
	err := yaml.Unmarshal(cfgBytes, cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *httpConfig) validate() error {
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
