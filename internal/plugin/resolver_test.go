package plugin

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
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
		Convey("Can resolve credentials", func() {
			resolver := newInstance()

			credentials := make([]*config_v2.Credential, 1, 1)
			credentials[0] = &config_v2.Credential{
				Name: "foo",
				From: "literal",
				Get:  "bar",
			}

			values, err := resolver.Resolve(credentials)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})

		Convey("Exits if credential resolution array is empty", func() {
			resolver := newInstance()

			credentials := make([]*config_v2.Credential, 0)

			resolveVarFunc := func() {
				resolver.Resolve(credentials)
			}

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if provider cannot be found", func() {
			resolver := newInstance()

			credentials := make([]*config_v2.Credential, 1, 1)
			credentials[0] = &config_v2.Credential{
				Name: "foo",
				From: "nope-not-found",
				Get:  "bar",
			}

			resolveVarFunc := func() {
				resolver.Resolve(credentials)
			}
			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if credential can't be resolved", func() {
			resolver := newInstance()

			credentials := make([]*config_v2.Credential, 1, 1)
			credentials[0] = &config_v2.Credential{
				Name: "foo",
				From: "env",
				Get:  "something-not-in-env",
			}

			credentialValues, err := resolver.Resolve(credentials)
			So(len(credentialValues), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			errorMsg := "ERROR: Resolving credentials from provider 'env' failed: env cannot find environment variable 'something-not-in-env'"
			So(err.Error(), ShouldEqual, errorMsg)

		})

		Convey("Can resolve credential2", func() {
			resolver := newInstance()

			credentials := make([]*config_v2.Credential, 1, 1)
			credentials[0] = &config_v2.Credential{
				Name: "foo",
				From: "literal",
				Get:  "bar",
			}

			values, err := resolver.Resolve(credentials)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})
	})
}
