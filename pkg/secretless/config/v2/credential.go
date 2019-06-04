package v2

import (
	"sort"

	"gopkg.in/yaml.v2"
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

	credentialBytes, err := yaml.Marshal(credYAML)
	if err != nil {
		return nil, err
	}

	// General Case: provider and id specified
	if _, ok := credYAML.(map[interface{}]interface{}); ok {
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

	// Special Case: string value
	} else {
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

	return cred, nil
}
