package plugin

import (
	"errors"
	"fmt"
	"testing"

	"github.com/conjurinc/secretless/internal/app/secretless/providers"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"

	. "github.com/smartystreets/goconvey/convey"
)

var fatalErrors []string

func newInstance() plugin_v1.Resolver {
	fatalErrors = []string{}
	logFatalf := func(format string, args ...interface{}) {
		fatalErrors = append(fatalErrors, fmt.Sprintf(format, args))
		panic(errors.New(fmt.Sprintf(format, args)))
	}

	return NewResolver(providers.ProviderFactories, nil, logFatalf)
}

func Test_Resolver(t *testing.T) {
	Convey("Resolve", t, func() {
		Convey("Can resolve variables", func() {
			resolver := newInstance()

			variables := make([]config.Variable, 1, 1)
			variables[0] = config.Variable{
				Name:     "foo",
				Provider: "literal",
				ID:       "bar",
			}

			values, err := resolver.Resolve(variables)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})

		Convey("Exits if variable resolution array is empty", func() {
			resolver := newInstance()

			variables := make([]config.Variable, 1, 1)

			resolveVarFunc := func() {
				resolver.Resolve(variables)
			}

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if provider cannot be found", func() {
			resolver := newInstance()

			variables := make([]config.Variable, 1, 1)
			variables[0] = config.Variable{
				Name:     "foo",
				Provider: "nope-not-found",
				ID:       "bar",
			}

			resolveVarFunc := func() {
				resolver.Resolve(variables)
			}
			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if variable can't be resolved", func() {
			resolver := newInstance()

			variables := make([]config.Variable, 1, 1)
			variables[0] = config.Variable{
				Name:     "foo",
				Provider: "env",
				ID:       "something-not-in-env",
			}

			variableValues, err := resolver.Resolve(variables)
			So(len(variableValues), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			errorMsg := "ERROR: Resolving variable 'something-not-in-env' from provider 'env' failed: env cannot find environment variable 'something-not-in-env'"
			So(err.Error(), ShouldEqual, errorMsg)

		})

		Convey("Can resolve variables2", func() {
			resolver := newInstance()

			variables := make([]config.Variable, 1, 1)
			variables[0] = config.Variable{
				Name:     "foo",
				Provider: "literal",
				ID:       "bar",
			}

			values, err := resolver.Resolve(variables)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)
		})
	})
}
