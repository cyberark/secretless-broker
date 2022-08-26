package plugin

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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
	t.Run("Exits if credential resolution array is empty", func(t *testing.T) {
		resolver := newInstance()
		credentials := make([]*config_v2.Credential, 0)
		resolveVarFunc := func() { resolver.Resolve(credentials) }

		assert.Panics(t, resolveVarFunc)
		assert.Len(t, fatalErrors, 1)
	})

	t.Run("Exits if even single provider cannot be found", func(t *testing.T) {
		resolver := newInstance()
		credentials := []*config_v2.Credential{
			{Name: "foo", From: "env", Get: "bar"},
			{Name: "foo", From: "nope-not-found", Get: "bar"},
			{Name: "baz", From: "also-not-found", Get: "bar"},
		}
		resolveVarFunc := func() { resolver.Resolve(credentials) }

		assert.Panics(t, resolveVarFunc)
		assert.Len(t, fatalErrors, 1)
	})

	t.Run("Returns an error if credential can't be resolved", func(t *testing.T) {
		resolver := newInstance()
		credentials := []*config_v2.Credential{
			{Name: "path", From: "env", Get: "PATH"},
			{Name: "foo", From: "env", Get: "something-not-in-env"},
			{Name: "bar", From: "env", Get: "something-also-not-in-env"},
			{Name: "baz", From: "file", Get: "something-not-on-file"},
		}
		credentialValues, err := resolver.Resolve(credentials)

		assert.Len(t, credentialValues, 0)
		assert.Error(t, err)
		assert.EqualError(t, err, "ERROR: Resolving credentials from provider 'env' failed: "+
			"env cannot find environment variable 'something-also-not-in-env', "+
			"env cannot find environment variable 'something-not-in-env'\n"+
			"ERROR: Resolving credentials from provider 'file' failed: "+
			"open something-not-on-file: no such file or directory")
	})

	t.Run("Can resolve credential", func(t *testing.T) {
		resolver := newInstance()
		credentials := []*config_v2.Credential{
			{Name: "foo", From: "env", Get: "PATH"},
			{Name: "bar", From: "literal", Get: "bar"},
		}
		values, err := resolver.Resolve(credentials)

		assert.NoError(t, err)
		assert.Len(t, values, 2)
		assert.Equal(t, os.Getenv("PATH"), string(values["foo"]))
		assert.Equal(t, "bar", string(values["bar"]))
	})
}
