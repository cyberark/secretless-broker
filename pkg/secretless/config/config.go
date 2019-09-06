package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"

	crd_api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// LoadFromFile loads a configuration file into a Config object.
func LoadFromFile(fileName string) (config config_v2.Config, err error) {
	var buffer []byte
	if buffer, err = ioutil.ReadFile(fileName); err != nil {
		err = fmt.Errorf("error reading config file '%s': '%s'", fileName, err)
		return
	}
	return Load(buffer)
}

// LoadFromCRD loads a configuration from a CRD API Configuration object
func LoadFromCRD(crdConfig crd_api_v1.Configuration) (config config_v2.Config, err error) {
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
func Load(data []byte) (config config_v2.Config, err error) {
	versionStruct := &struct {
		Version string `yaml:"version"`
	}{}

	if err = yaml.Unmarshal(data, versionStruct); err != nil {
		err = fmt.Errorf("unable to load configuration: '%s'", err)
		return
	}

	var configPointer *config_v2.Config
	if versionStruct.Version == "" {
		versionStruct.Version = "1"
	}

	switch versionStruct.Version {
	case "1":
		log.Printf("WARN: v1 configuration is now deprecated and will be removed in a future release")

		var v1Config *config_v1.Config
		if v1Config, err = config_v1.NewConfig(data); err != nil {
			err = fmt.Errorf("unable to load configuration when parsing version 1: '%s'", err)
		}

		if configPointer, err = config_v1.NewV2Config(v1Config); err != nil {
			err = fmt.Errorf("unable to load configuration when parsing version 1: '%s'", err)
		}
	case "2":
		if configPointer, err = config_v2.NewConfig(data); err != nil {
			err = fmt.Errorf("unable to load configuration when parsing version 2: '%s'", err)
		}
	default:
		err = fmt.Errorf("unknown configuration version '%s'", versionStruct.Version)
	}

	if err != nil {
		return
	}
	config = *configPointer

	return
}
