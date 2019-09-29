package v2

import (
	"sort"

	"gopkg.in/yaml.v2"
)

// Credential is the v2 representation of a named secret stored in a provider.
// It's the analog of the v1.StoredSecret.
// TODO: Move to types file along with other non-dependency types.
type Credential struct {
	Name string
	From string
	Get  string
}

// NewCredentials converts the raw YAML representation of credentials
// (credentialsYAML) into it's logical representation ([]*Credential).
func NewCredentials(credsYAML credentialsYAML) ([]*Credential, error) {
	credentials := make([]*Credential, 0)

	for credName, credYAML := range credsYAML {
		cred, err := NewCredential(credName, credYAML)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	// Sort credentials
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	return credentials, nil
}

// NewCredential creates a Credential from a credential name and raw yaml
// that's been unmarshalled into an interface{}.
func NewCredential(credName string, credYAML interface{}) (*Credential, error) {
	cred := &Credential{
		Name: credName,
	}

	credentialBytes, err := yaml.Marshal(credYAML)
	if err != nil {
		return nil, err
	}

	// Special Case: scalar literal value specified
	if _, ok := credYAML.(map[interface{}]interface{}); !ok {
		// Try to unmarshall it into a string
		var credentialValue string
		err = yaml.Unmarshal(credentialBytes, &credentialValue)

		// TODO: the line number in this error is dishonest
		if err != nil {
			return nil, err
		}

		cred.From = "literal"
		cred.Get = credentialValue
		return cred, nil
	}

	// General Case: provider and id specified
	credYamlStruct := &struct {
		From string `yaml:"from"`
		Get  string `yaml:"get"`
	}{}

	err = yaml.Unmarshal(credentialBytes, credYamlStruct)

	// TODO: the line number in this error is dishonest
	if err != nil {
		return nil, err
	}

	cred.Get = credYamlStruct.Get
	cred.From = credYamlStruct.From

	return cred, nil
}
