package v2

import validation "github.com/go-ozzo/ozzo-validation"

// ConnectorConfig is a wrapper around byte slice
// that allows the connector configuration
// to be Marshalled to YAML.
type ConnectorConfig []byte

func (c ConnectorConfig) MarshalYAML() (interface{}, error) {
	return string(c), nil
}

func (c ConnectorConfig) Bytes() []byte {
	return c
}

// Service represents a the configuration of a Secretless proxy service. It
// includes the service's protocol, the socket or address it listens on, the
// location of its required credentials, and (optionally) any additional
// protocol specific configuration.
type Service struct {
	Debug 			bool
	Connector       string
	ConnectorConfig ConnectorConfig
	Credentials     []*Credential
	ListenOn        string
	Name            string
}

// HasCredential indicates whether a Service has the specified credential.
func (s Service) HasCredential(credentialName string) bool {
	for _, credential := range s.Credentials {
		if credential.Name == credentialName {
			return true
		}
	}
	return false
}

// Validate verifies the completeness and correctness of the Service.
func (s Service) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Credentials, validation.Required),
		validation.Field(&s.ListenOn, validation.Required),
		validation.Field(&s.Name, validation.Required),
	)
}
