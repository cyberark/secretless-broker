package config

import (
	"fmt"
	"io/ioutil"
	"regexp"

	crdAPIv1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config/config_v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config/config_v2"
	yaml "gopkg.in/yaml.v2"
)

// LoadFromFile loads a configuration file into a Config object.
func LoadFromFile(fileName string) (config config_v1.Config, err error) {
	var buffer []byte
	if buffer, err = ioutil.ReadFile(fileName); err != nil {
		err = fmt.Errorf("error reading config file '%s': '%s'", fileName, err)
		return
	}
	return Load(buffer)
}

// LoadFromCRD loads a configuration from a CRD API Configuration object
func LoadFromCRD(crdConfig crdAPIv1.Configuration) (config config_v1.Config, err error) {
	var specData []byte
	if specData, err = yaml.Marshal(crdConfig.Spec); err != nil {
		return
	}

	if config, err = Load(specData); err != nil {
		return
	}

	return config, nil
}

// Load parses a YAML string into a Config object.
func Load(data []byte) (config config_v1.Config, err error) {
	versionStruct := &struct {
		Version string `yaml:"version"`
	}{}

	if err = yaml.Unmarshal(data, versionStruct); err != nil {
		err = fmt.Errorf("unable to load configuration: '%s'", err)
		return
	}

	var configPointer *config_v1.Config
	if versionStruct.Version == "" {
		versionStruct.Version = "1"
	}

	switch versionStruct.Version {
	case "1":
		if configPointer, err = config_v1.NewConfig(data); err != nil {
			err = fmt.Errorf("unable to load configuration when parsing version 1: '%s'", err)
		}
	case "2":
		if configPointer, err = config_v2.NewV1Config(data); err != nil {
			err = fmt.Errorf("unable to load configuration when parsing version 2: '%s'", err)
		}
	default:
		err = fmt.Errorf("unknown configuration version '%s'", versionStruct.Version)
	}

	if err != nil {
		return
	}
	config = *configPointer

	for i := range config.Listeners {
		l := &config.Listeners[i]
		if l.Protocol == "" {
			l.Protocol = l.Name
		}
	}

	for i := range config.Handlers {
		h := &config.Handlers[i]
		if h.Type == "" {
			h.Type = h.Name
		}
		if h.ListenerName == "" {
			h.ListenerName = h.Name
		}

		h.Patterns = make([]*regexp.Regexp, len(h.Match))
		for i, matchPattern := range h.Match {
			pattern, err := regexp.Compile(matchPattern)
			if err != nil {
				panic(err.Error())
			} else {
				h.Patterns[i] = pattern
			}
		}
	}

	if err = config.Validate(); err != nil {
		err = fmt.Errorf("configuration is not valid: %s", err)
		return
	}

	return
}
