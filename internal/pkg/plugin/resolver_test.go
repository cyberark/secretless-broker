package plugin

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/app/secretless/providers"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

var fatalErrors []string

func newInstance() plugin_v1.Resolver {
	fatalErrors = []string{}
	logFatalf := func(format string, args ...interface{}) {
		fatalErrors = append(fatalErrors, fmt.Sprintf(format, args))
		panic(fmt.Errorf(format, args))
	}

	return NewResolver(providers.ProviderFactories, nil, logFatalf)
}

func Test_Resolver(t *testing.T) {
	Convey("Resolve", t, func() {
		Convey("Can resolve secrets", func() {
			resolver := newInstance()

			secrets := make([]*config_v2.Credential, 1, 1)
			secrets[0] = &config_v2.Credential{
				Name: "foo",
				From: "literal",
				Get:  "bar",
			}

			values, err := resolver.Resolve(secrets)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})

		Convey("Exits if secret resolution array is empty", func() {
			resolver := newInstance()

			secrets := make([]*config_v2.Credential, 0)

			resolveVarFunc := func() {
				resolver.Resolve(secrets)
			}

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if provider cannot be found", func() {
			resolver := newInstance()

			secrets := make([]*config_v2.Credential, 1, 1)
			secrets[0] = &config_v2.Credential{
				Name: "foo",
				From: "nope-not-found",
				Get:  "bar",
			}

			resolveVarFunc := func() {
				resolver.Resolve(secrets)
			}
			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if secret can't be resolved", func() {
			resolver := newInstance()

			secrets := make([]*config_v2.Credential, 1, 1)
			secrets[0] = &config_v2.Credential{
				Name: "foo",
				From: "env",
				Get:  "something-not-in-env",
			}

			secretValues, err := resolver.Resolve(secrets)
			So(len(secretValues), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			errorMsg := "ERROR: Resolving variable 'something-not-in-env' from provider 'env' failed: env cannot find environment variable 'something-not-in-env'"
			So(err.Error(), ShouldEqual, errorMsg)

		})

		Convey("Can resolve secret2", func() {
			resolver := newInstance()

			secrets := make([]*config_v2.Credential, 1, 1)
			secrets[0] = &config_v2.Credential{
				Name: "foo",
				From: "literal",
				Get:  "bar",
			}

			values, err := resolver.Resolve(secrets)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})
	})
}
