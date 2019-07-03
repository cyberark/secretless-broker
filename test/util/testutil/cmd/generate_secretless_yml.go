package main

import (
	"io/ioutil"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	"gopkg.in/yaml.v2"
)

func main()  {
	secretlessConfig, _ := testutil.GenerateConfigurations()
	d, err := yaml.Marshal(&secretlessConfig)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("./fixtures/secretless.yml", d, 0644)
}
