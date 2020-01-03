package v2

import (
	"errors"
	"regexp"

	"github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

type httpConfigYAML struct {
	AuthenticateURLsMatching []string `yaml:"authenticateURLsMatching"`
}

// HTTPConfig represents service-specific configuration for service connectors
// built on top of the http protocol
type HTTPConfig struct {
	AuthenticateURLsMatching []*regexp.Regexp
}

// HTTPAuthenticationStrategies are the different ways an http service
// can authenticate.
var HTTPAuthenticationStrategies = []interface{}{
	"aws",
	"basic_auth",
	"conjur",
}

// IsHTTPConnector returns true iff the connector provided
// uses the http protocol
func IsHTTPConnector(connector string) bool {
	for _, strategy := range HTTPAuthenticationStrategies {
		if strategy == connector {
			return true
		}
	}
	return false
}

func newHTTPConfigYAML(cfgBytes []byte) (*httpConfigYAML, error) {
	cfg := &httpConfigYAML{}
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

// NewHTTPConfig creates an HTTPConfig from yaml bytes
func NewHTTPConfig(cfgBytes []byte) (*HTTPConfig, error) {
	cfg, err := newHTTPConfigYAML(cfgBytes)
	if err != nil {
		return nil, err
	}

	AuthenticateURLsMatching := make([]*regexp.Regexp, len(cfg.AuthenticateURLsMatching))
	for i, matchPattern := range cfg.AuthenticateURLsMatching {
		pattern, err := regexp.Compile(matchPattern)
		if err != nil {
			panic(err.Error())
		} else {
			AuthenticateURLsMatching[i] = pattern
		}
	}

	return &HTTPConfig{
		AuthenticateURLsMatching: AuthenticateURLsMatching,
	}, nil
}

func (cfg *httpConfigYAML) UnmarshalYAML(bytes []byte) error {
	// Unmarshall into a temp struct
	//
	// This temp struct makes it possible to parse 'authenticateURLsMatching' as
	// string or []string
	tempCfg := &struct {
		AuthenticateURLsMatching interface{} `yaml:"authenticateURLsMatching"`
	}{}
	err := yaml.Unmarshal(bytes, tempCfg)
	if err != nil {
		return errors.New("http connectorConfig could not be parsed")
	}

	// Populate actual http config from tempCfg
	switch v := tempCfg.AuthenticateURLsMatching.(type) {
	case string:
		cfg.AuthenticateURLsMatching = []string{v}
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
func (cfg *httpConfigYAML) validate() error {
	return validation.ValidateStruct(
		cfg,
		// AuthenticateURLsMatching cannot be empty
		validation.Field(&cfg.AuthenticateURLsMatching, validation.Required),
	)
}
