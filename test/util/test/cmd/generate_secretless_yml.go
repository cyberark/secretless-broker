package main

import (
	. "github.com/cyberark/secretless-broker/test/util/test"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func main()  {
	secretlessConfig, _ := GenerateConfigurations()
	d, err := yaml.Marshal(&secretlessConfig)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("./fixtures/secretless.yml", d, 0644)
}
