package v2

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

type httpConfig struct {
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

// HTTPAuthenticationStrategies are the different ways an http service
// can authenticate.
var HTTPAuthenticationStrategies = []interface{}{
	"aws",
	"basic_auth",
	"conjur",
}

func isHTTPConnector(connector string) bool {
	for _, strategy := range HTTPAuthenticationStrategies {
		if strategy == connector {
			return true
		}
	}
	return false
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
	// Unmarshall into a temp struct
	//
	// This temp struct makes it possible to parse 'authenticateURLsMatching' as
	// string or []string
	tempCfg := &struct {
		AuthenticateURLsMatching interface{} `yaml:"authenticateURLsMatching"`
	}{}
	err := yaml.Unmarshal(bytes, tempCfg)
	if err != nil {
		return errors.New("http ConnectorConfig could not be parsed")
	}

	// Populate actual http config from tempCfg
	switch v := tempCfg.AuthenticateURLsMatching.(type) {
	case string:
		cfg.AuthenticateURLsMatching = []string{ v }
	case []interface{}:
		urlMatchStrings := make([]string, len(v))
		for i, urlMatch := range v {
			urlMatchString, ok := urlMatch.(string)
			if !ok {
				return errors.New("'authenticateURLsMatching' key has incorrect type, must be a string or list of strings")
			}
			urlMatchStrings[i] = urlMatchString
		}

		cfg.AuthenticateURLsMatching = urlMatchStrings
	default:
		return errors.New("'authenticateURLsMatching' key has incorrect type, must be a string or list of strings")
	}

	return nil
}

// validate carries out validation of httpConfig
// ensuring that the validation rules of fields are met
func (cfg *httpConfig) validate() error {
	return validation.ValidateStruct(
		cfg,
		// AuthenticateURLsMatching cannot be empty
		validation.Field(&cfg.AuthenticateURLsMatching, validation.Required),
	)
}
