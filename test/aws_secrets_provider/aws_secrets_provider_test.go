package main

import (
	"testing"

	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

func TestAWSSecrets_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider

	name := "aws"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	Convey("Can create the AWS Secrets provider", t, func() {
		provider, err = providers.ProviderFactories[name](options)
		So(err, ShouldBeNil)
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, "aws")
	})
}
