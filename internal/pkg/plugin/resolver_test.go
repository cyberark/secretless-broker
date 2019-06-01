package plugin

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"testing"

	"github.com/cyberark/secretless-broker/internal/app/secretless/providers"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"

	. "github.com/smartystreets/goconvey/convey"
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

			secrets := make([]v1.StoredSecret, 1, 1)
			secrets[0] = v1.StoredSecret{
				Name:     "foo",
				Provider: "literal",
				ID:       "bar",
			}

			values, err := resolver.Resolve(secrets)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})

		Convey("Exits if secret resolution array is empty", func() {
			resolver := newInstance()

			secrets := make([]v1.StoredSecret, 1, 1)

			resolveVarFunc := func() {
				resolver.Resolve(secrets)
			}

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if provider cannot be found", func() {
			resolver := newInstance()

			secrets := make([]v1.StoredSecret, 1, 1)
			secrets[0] = v1.StoredSecret{
				Name:     "foo",
				Provider: "nope-not-found",
				ID:       "bar",
			}

			resolveVarFunc := func() {
				resolver.Resolve(secrets)
			}
			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if secret can't be resolved", func() {
			resolver := newInstance()

			secrets := make([]v1.StoredSecret, 1, 1)
			secrets[0] = v1.StoredSecret{
				Name:     "foo",
				Provider: "env",
				ID:       "something-not-in-env",
			}

			secretValues, err := resolver.Resolve(secrets)
			So(len(secretValues), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			errorMsg := "ERROR: Resolving variable 'something-not-in-env' from provider 'env' failed: env cannot find environment variable 'something-not-in-env'"
			So(err.Error(), ShouldEqual, errorMsg)

		})

		Convey("Can resolve secret2", func() {
			resolver := newInstance()

			secrets := make([]v1.StoredSecret, 1, 1)
			secrets[0] = v1.StoredSecret{
				Name:     "foo",
				Provider: "literal",
				ID:       "bar",
			}

			values, err := resolver.Resolve(secrets)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})
	})
}
