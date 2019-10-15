package sharedobj

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj/mock"
)

func TestCompatiblePluginVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", CompatiblePluginAPIVersion)
}

// TestPlugins uses the actual implementation of "Plugins" -- because that's
// what we're testing.  We populate "Plugins" with mocks of the plugin instances
// themselves, because those don't matter: we only care that we get back what we
// put in. That is, this is simply a test that the method HTTPPlugins() returns
// what we give it.
func TestPlugins(t *testing.T) {
	t.Run("HTTPPlugins() returns expected plugins", func(t *testing.T) {

		plugins := &Plugins{
			HTTPPluginsByID: mock.HTTPPluginsById,
			TCPPluginsByID:  mock.TCPPluginsById,
		}

		httpPlugins := plugins.HTTPPlugins()

		assert.NotNil(t, httpPlugins)
		if httpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, httpPlugins, mock.HTTPPluginsById)
	})

	t.Run("TCPPlugins() returns expected plugins", func(t *testing.T) {
		plugins := &Plugins{
			HTTPPluginsByID: mock.HTTPPluginsById,
			TCPPluginsByID:  mock.TCPPluginsById,
		}

		tcpPlugins := plugins.TCPPlugins()

		assert.NotNil(t, tcpPlugins)
		if tcpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, tcpPlugins, mock.TCPPluginsById)
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
			t.Fail()
		}

		assert.EqualValues(t, mock.AllHTTPPlugins(), allPlugins.HTTPPlugins())
		assert.EqualValues(t, mock.AllTCPPlugins(), allPlugins.TCPPlugins())
	})

	t.Run("External plugins override same-named internal plugins", func(t *testing.T) {

		// Setup our mocks

		extHTTPPlugins := mock.ExternalPlugins().HTTPPlugins()
		extHTTPPlugins["intHTTP2"] = mock.HTTPPlugin{}

		extTCPPlugins := mock.ExternalPlugins().TCPPlugins()
		extTCPPlugins["intTCP1"] = mock.TCPPlugin{}

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
			t.Fail()
		}

		// Ensure the override occurred
		assert.Equal(t, extHTTPPlugins["intHTTP2"], allPlugins.HTTPPlugins()["intHTTP2"])
		assert.Equal(t, extTCPPlugins["intTCP1"], allPlugins.TCPPlugins()["intTCP1"])

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
