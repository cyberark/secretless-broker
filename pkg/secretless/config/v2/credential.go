package v2

import (
	"gopkg.in/yaml.v2"
)

type Credential struct {
	Name string
	From string
	Get  string
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

