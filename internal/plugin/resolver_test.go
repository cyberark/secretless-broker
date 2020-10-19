package plugin

import (
	"fmt"
	"os"
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
		Convey("Exits if credential resolution array is empty", func() {
			resolver := newInstance()
			credentials := make([]*config_v2.Credential, 0)
			resolveVarFunc := func() { resolver.Resolve(credentials) }

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Exits if even single provider cannot be found", func() {
			resolver := newInstance()
			credentials := []*config_v2.Credential{
				{Name: "foo", From: "env", Get: "bar"},
				{Name: "foo", From: "nope-not-found", Get: "bar"},
				{Name: "baz", From: "also-not-found", Get: "bar"},
			}
			resolveVarFunc := func() { resolver.Resolve(credentials) }

			So(resolveVarFunc, ShouldPanic)
			So(len(fatalErrors), ShouldEqual, 1)
		})

		Convey("Returns an error if credential can't be resolved", func() {
			resolver := newInstance()
			credentials := []*config_v2.Credential{
				{Name: "path", From: "env", Get: "PATH"},
				{Name: "foo", From: "env", Get: "something-not-in-env"},
				{Name: "bar", From: "env", Get: "something-also-not-in-env"},
				{Name: "baz", From: "file", Get: "something-not-on-file"},
			}
			credentialValues, err := resolver.Resolve(credentials)

			So(len(credentialValues), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual,
				"ERROR: Resolving credentials from provider 'env' failed: "+
					"env cannot find environment variable 'something-also-not-in-env', "+
					"env cannot find environment variable 'something-not-in-env'\n"+
					"ERROR: Resolving credentials from provider 'file' failed: "+
					"open something-not-on-file: no such file or directory")
		})

		Convey("Can resolve credential", func() {
			resolver := newInstance()
			credentials := []*config_v2.Credential{
				{Name: "foo", From: "env", Get: "PATH"},
				{Name: "bar", From: "literal", Get: "bar"},
			}
			values, err := resolver.Resolve(credentials)

			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 2)
			So(string(values["foo"]), ShouldEqual, os.Getenv("PATH"))
			So(string(values["bar"]), ShouldEqual, "bar")
		})
	})
}
