package v2

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

// Config represents a full configuration of Secretless, which is just a list of
// individual Service configurations.
type Config struct {
	Debug bool
	Services []*Service
}

// Validate verifies the completeness and correctness of the Config.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Services, validation.Required),
	)
}

// Serialize Config to YAML
func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewConfig creates a v2.Config from yaml bytes
func NewConfig(v2YAML []byte) (*Config, error) {
	cfgYAML, err := newConfigYAML(v2YAML)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0)
	for svcName, svcYAML := range cfgYAML.Services {
		svc, err := NewService(svcName, svcYAML)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	// sort Services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return &Config{
		Services: services,
	}, nil
}
