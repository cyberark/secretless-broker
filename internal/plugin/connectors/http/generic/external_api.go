package generic

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// ConfigYAML is an almost-literal representation of the YAML config section
// of a generic HTTP connector.  It has two purposes:
//
//     1.  It allows other internal connectors to configure a generic connector
//         to which they are delegating their implementation.  That is, it's
//         a code-level interface equivalent of a .yml file config section.
//     2.  Technically, it acts as an intermediate form when parsing an actual
//         .yml file.  That is, we convert from .yml --> ConfigYAML --> config.
type ConfigYAML struct {
	CredentialValidations map[string]string `yaml:"credentialValidations"`
	ForceSSL              bool              `yaml:"forceSSL"`
	Headers               map[string]string `yaml:"headers"`
	OAuth1Secrets         map[string]string `yaml:"oauth1"`
	QueryParams           map[string]string `yaml:"queryParams"`
}

// NewConfigYAML takes the raw cfgBytes and unmarshals them into a ConfigYAML
// struct.
func NewConfigYAML(cfgBytes []byte) (*ConfigYAML, error) {
	cfgYAML := &ConfigYAML{
		// TODO: Fix CRDs so we can make the default true
		ForceSSL: false, // Default should be false
	}

	err := yaml.Unmarshal(cfgBytes, cfgYAML)
	if err != nil {
		return nil, err
	}

	return cfgYAML, nil
}

// NewConnectorConstructor is a meta-constructor: a function that returns a
// NewConnector function. It's intended to be used by other HTTP connectors who
// delegate their implementation to the generic HTTP connector.
//
// It takes a static ConfigYAML -- static in the sense that its represents the
// connector definition, not user configuration -- which is then merged at
// runtime with the actual config from the .yml file, allowing ForceSSL to be
// configured by the user.
func NewConnectorConstructor(staticCfgYAML *ConfigYAML) (http.NewConnectorFunc, error) {

	staticCfg, err := newConfig(staticCfgYAML)
	if err != nil {
		return nil, fmt.Errorf("invalid ConfigYAML: %s", err)
	}

	return func(conRes connector.Resources) http.Connector {
		logger := conRes.Logger()

		// Allow runtime config to override ForceSSL
		runtimeCfgYAML, err := NewConfigYAML(conRes.Config())
		if err != nil {
			logger.Panicf("can't create connector: can't unmarshal YAML.")
		}

		// Don't have to worry about nil condition because of the panic:
		//noinspection GoNilness
		staticCfg.ForceSSL = runtimeCfgYAML.ForceSSL

		return &Connector{
			logger: logger,
			config: staticCfg,
		}
	}, nil
}
