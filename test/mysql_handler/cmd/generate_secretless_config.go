package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	. "github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
)

func main()  {
	secretlessConfig, _ := GenerateConfigurations()
	d, err := yaml.Marshal(&secretlessConfig)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("secretless.yml", d, 0644)
}
