package v2

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

type httpConfig struct {
	AuthenticationStrategy   string   `yaml:"authenticationStrategy"`
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

// HttpAuthenticationStrategies are the different ways an http service
// can authenticate.
var HttpAuthenticationStrategies = []string{
	"aws",
	"basic_auth",
	"conjur",
}

func newHTTPConfig(cfgBytes []byte) (*httpConfig, error) {
	cfg := &httpConfig{}
	err := cfg.UnmarshalYAML(cfgBytes)
	if err != nil {
		return nil, err
	}

	err = cfg.validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *httpConfig) UnmarshalYAML(bytes []byte) error {
	err := yaml.Unmarshal(bytes, cfg)

	// A string type is converted into [] by default, so we must verify length.
	// If this passes, all is good and we can return.
	if ok := err == nil && len(cfg.AuthenticateURLsMatching) > 0; ok {
		return nil
	}

	// Unmarshall into a temp struct that expects a string
	tempCfg := &struct {
		AuthenticationStrategy   string `yaml:"authenticationStrategy"`
		AuthenticateURLsMatching string `yaml:"authenticateURLsMatching"`
	}{}
	err = yaml.Unmarshal(bytes, tempCfg)

	// It must succeed with a non-empty string
	if ok := err == nil && len(tempCfg.AuthenticateURLsMatching) > 0; !ok {
		return errors.New("http ProtocolConfig could not be parsed")
	}

	// It's a string, let's convert it to a []string
	cfg.AuthenticationStrategy = tempCfg.AuthenticationStrategy
	cfg.AuthenticateURLsMatching = []string{ tempCfg.AuthenticateURLsMatching }

	return nil
}

// validate ensures AuthenticationStrategy neither field is empty, and that
// AuthentcationStrategy is a valid value
func (cfg *httpConfig) validate() error {
	// convert strategies from []string to []interface{} for validation.In
	var availStrategies []interface{}
	for _, s := range HttpAuthenticationStrategies {
		availStrategies = append(availStrategies, s)
	}

	return validation.ValidateStruct(
		cfg,
		validation.Field(
			&cfg.AuthenticationStrategy,
			validation.Required,
			validation.In(availStrategies...),
		),
		// AuthenticateURLsMatching cannot be empty
		validation.Field(&cfg.AuthenticateURLsMatching, validation.Required),
	)
}
