package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/cyberark/secretless-broker/test/util/testutil"
)

func main() {
	secretlessConfig, _ := testutil.GenerateConfigurations()
	d, err := yaml.Marshal(secretlessConfig)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("./fixtures/secretless.yml", d, 0644)
}
