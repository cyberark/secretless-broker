package v2

import (
	"gopkg.in/yaml.v2"
	"sort"
)

// Credential the v2.Config representation of a named secret stored in a
// provider. It's the analog of the StoredSecret in v1.Config.
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

	// sort credentials
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	return credentials, nil
}

func NewCredential(credName string, credYAML interface{}) (*Credential, error) {

	cred := &Credential{
		Name: credName,
	}

	// Special Case: string value

	if credVal, ok := credYAML.(string); ok {
		cred.From = "literal"
		cred.Get = credVal
		return cred, nil
	}

	// General Case: provider and id specified

	credentialBytes, err := yaml.Marshal(credYAML)
	if err != nil {
		return nil, err
	}

	credYamlStruct := &struct {
		From string `yaml:"from"`
		Get  string `yaml:"get"`
	}{}

	err = yaml.Unmarshal(credentialBytes, credYamlStruct)
	if err != nil {
		return nil, err
	}

	cred.Get = credYamlStruct.Get
	cred.From = credYamlStruct.From

	return cred, nil
}
