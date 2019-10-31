package sharedobj

import (
	"errors"
	"testing"

	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj/mock"
)

func TestPluginAPIVersionIsSemverString(t *testing.T) {
	semVerRe := `\d+\.\d+\.\d+`
	assert.Regexp(t, semVerRe, CompatiblePluginAPIVersion)
}

// TestPlugins uses the actual implementation of "Plugins" -- because that's
// what we're testing.  We populate "Plugins" with mocks of the plugin instances
// themselves, because those don't matter: we only care that we get back what we
// put in. That is, this is simply a test that the methods HTTPPlugins() and
// TCPPlugins() return what we give them.
func TestPlugins(t *testing.T) {
	t.Run("HTTPPlugins() and TCPPlugins() return expected plugins", func(t *testing.T) {

		expectedHTTP := mock.HTTPInternalPluginsByID()
		expectedTCP := mock.TCPInternalPluginsByID()

		plugins := &Plugins{
			HTTPPluginsByID: expectedHTTP,
			TCPPluginsByID:  expectedTCP,
		}

		// Test HTTPPlugins()

		actualHTTP := plugins.HTTPPlugins()

		assert.NotNil(t, actualHTTP)
		if actualHTTP == nil {
			return
		}

		assert.Equal(t, expectedHTTP, actualHTTP)

		// Test TCPPlugins()

		actualTCP := plugins.TCPPlugins()

		assert.NotNil(t, actualTCP)
		if actualTCP == nil {
			return
		}

		assert.Equal(t, expectedTCP, actualTCP)
	})
}

func TestAllAvailablePlugins(t *testing.T) {
	t.Run("Assembles internal and external plugins", func(t *testing.T) {
		allPlugins, err := AllAvailablePluginsWithOptions(
			"",
			"",
			mock.GetInternalPlugins,
			mock.GetExternalPlugins,
			mock.NewLogger(),
		)
		assert.Nil(t, err)
		if err != nil {
			return
		}

		assert.EqualValues(t, mock.AllHTTPPlugins(), allPlugins.HTTPPlugins())
		assert.EqualValues(t, mock.AllTCPPlugins(), allPlugins.TCPPlugins())
	})

	t.Run("External plugins override same-named internal plugins", func(t *testing.T) {
		// Setup our mocks
		httpOverride := mock.NewHTTPPlugin("HTTP Override")
		extHTTPPlugins := mock.ExternalPlugins().HTTPPlugins()
		extHTTPPlugins["intHTTP2"] = httpOverride // Override key "intHTTP2"

		tcpOverride := mock.NewTCPPlugin("TCP Override")
		extTCPPlugins := mock.ExternalPlugins().TCPPlugins()
		extTCPPlugins["intTCP1"] = tcpOverride // Override key "intTCP1"

		allExternalPlugins := Plugins{
			HTTPPluginsByID: extHTTPPlugins,
			TCPPluginsByID:  extTCPPlugins,
		}

		getExternalPlugins := func(
			pluginDir string,
			checksumfile string,
			logger log_api.Logger,
		) (plugin.AvailablePlugins, error) {
			return &allExternalPlugins, nil
		}

		// Create the test subject

		allPlugins, err := AllAvailablePluginsWithOptions(
			"",
			"",
			mock.GetInternalPlugins,
			getExternalPlugins,
			mock.NewLogger(),
		)

		// Test

		assert.Nil(t, err)
		if err != nil {
			return
		}


		// Define our expected results (without override)
		expectedHTTP := mock.AllHTTPPlugins()
		expectedTCP := mock.AllTCPPlugins()

		// Sanity check: Not equal without the override
		assert.NotEqual(t, expectedHTTP, allPlugins.HTTPPlugins())
		assert.NotEqual(t, expectedTCP, allPlugins.TCPPlugins())

		// Now do the override
		expectedHTTP["intHTTP2"] = httpOverride
		expectedTCP["intTCP1"] = tcpOverride

		// Now it should be equal
		assert.Equal(t, expectedHTTP, allPlugins.HTTPPlugins(), "HTTP override failed")
		assert.Equal(t, expectedTCP, allPlugins.TCPPlugins(), "TCP override failed")
	})

	t.Run("GetExternalPlugins receives arguments and propagates errors", func(t *testing.T) {
		expectedDir := "fooDir"
		expectedChecksumfile := "fooChecksum"
		expectedLogger := mock.NewLogger()
		expectedErr := errors.New("foo")

		var getExternalPlugins = func(
			pluginDir string,
			checksumfile string,
			logger log_api.Logger,
		) (plugin.AvailablePlugins, error) {
			assert.Equal(t, expectedDir, pluginDir)
			assert.Equal(t, expectedChecksumfile, checksumfile)
			assert.Equal(t, expectedLogger, logger)
			return nil, expectedErr
		}

		_, actualErr := AllAvailablePluginsWithOptions(
			expectedDir,
			expectedChecksumfile,
			mock.GetInternalPlugins,
			getExternalPlugins,
			expectedLogger,
		)

		// NOTE: This assertion is needed to ensure the other assertions were run
		assert.Equal(t, expectedErr, actualErr)
	})

	t.Run("GetInternalPlugins propagates errors", func(t *testing.T) {
		expectedErr := errors.New("mock InternalPluginLookupFunc error")

		var getInternalPlugins = func() (plugin.AvailablePlugins, error) {
			return nil, expectedErr
		}

		_, err := AllAvailablePluginsWithOptions(
			"directory",
			"hashfile",
			getInternalPlugins,
			mock.GetExternalPlugins,
			mock.NewLogger(),
		)

		assert.Equal(t, expectedErr, err)
	})
}
