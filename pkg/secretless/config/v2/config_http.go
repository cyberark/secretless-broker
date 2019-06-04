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

// ValidHttpAuthenticationStrategies is a []interface rather than a []string
// because the validation method expects that
var ValidHttpAuthenticationStrategies = []interface{}{
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

	// string is converted into [] by default, so we must verify length
	ok := err == nil && len(cfg.AuthenticateURLsMatching) > 0

	// it worked, just return
	if ok {
		return nil
	}

	// it failed, let's check if authenticateURLsMatching is a string

	// unmarshall into a temp struct that expects a string
	tempCfg := &struct {
		AuthenticationStrategy   string `yaml:"authenticationStrategy"`
		AuthenticateURLsMatching string `yaml:"authenticateURLsMatching"`
	}{}
	err = yaml.Unmarshal(bytes, tempCfg)

	// it must succeed with a non-empty string
	ok = err == nil && len(tempCfg.AuthenticateURLsMatching) > 0

	// still failed, this is a real error and not a valid string pattern
	if !ok {
		return errors.New("http ProtocolConfig could not be parsed")
	}

	// it's a string, let's convert it to a []string
	cfg.AuthenticationStrategy = tempCfg.AuthenticationStrategy
	cfg.AuthenticateURLsMatching = []string{ tempCfg.AuthenticateURLsMatching }

	return nil
}

// validate ensures AuthenticationStrategy neither field is empty, and that
// AuthentcationStrategy is a valid value
func (cfg *httpConfig) validate() error {
	return validation.ValidateStruct(
		cfg,
		validation.Field(
			&cfg.AuthenticationStrategy,
			validation.Required,
			validation.In(ValidHttpAuthenticationStrategies...),
			//validation.In("aws", "basic_auth", "conjur"),
		),
		// AuthenticateURLsMatching cannot be empty
		validation.Field(&cfg.AuthenticateURLsMatching, validation.Required),
	)
}
