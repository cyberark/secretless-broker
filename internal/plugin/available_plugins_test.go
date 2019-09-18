package plugin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/log"
	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

type mockHTTPPlugin struct{ http.Plugin }
type mockTCPPlugin struct{ tcp.Plugin }

// These need to sit outside of the func since we are comparing them in the
// tests
var mockHTTPPlugins = map[string]http.Plugin{
	"one": &mockHTTPPlugin{},
	"two": &mockHTTPPlugin{},
}

var mockTCPPlugins = map[string]tcp.Plugin{
	"one":   &mockTCPPlugin{},
	"two":   &mockTCPPlugin{},
	"three": &mockTCPPlugin{},
}

func mockPlugins() AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: mockHTTPPlugins,
		TCPPluginsByID:  mockTCPPlugins,
	}
}

func mockInternalPlugins() AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"intHTTP1": &mockHTTPPlugin{},
			"intHTTP2": &mockHTTPPlugin{},
			"intHTTP3": &mockHTTPPlugin{},
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"intTCP1": &mockTCPPlugin{},
			"intTCP2": &mockTCPPlugin{},
			"intTCP3": &mockTCPPlugin{},
		},
	}
}

func mockGetInternalPlugins() (AvailablePlugins, error) {
	return mockInternalPlugins(), nil
}

func mockExternalPlugins() AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"extHTTP1": &mockHTTPPlugin{},
			"extHTTP2": &mockHTTPPlugin{},
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"extTCP1": &mockTCPPlugin{},
			"extTCP2": &mockTCPPlugin{},
			"extTCP3": &mockTCPPlugin{},
		},
	}
}

func mockGetExternalPlugins(
	pluginDir string,
	checksumfile string,
	logger log_api.Logger,
) (AvailablePlugins, error) {

	return mockExternalPlugins(), nil
}

func assertMapContainsHTTPPlugins(
	t *testing.T,
	allPlugins AvailablePlugins,
	expectedPlugins map[string]http.Plugin,
	comparePointers bool,
) {

	for name, httpPlugin := range expectedPlugins {
		assert.Contains(t, allPlugins.HTTPPlugins(), name)
		if _, ok := allPlugins.HTTPPlugins()[name]; !ok {
			t.Fail()
		}

		if comparePointers {
			assert.Equal(t, httpPlugin, allPlugins.HTTPPlugins()[name])
		}
	}

}

func assertMapContainsTCPPlugins(
	t *testing.T,
	allPlugins AvailablePlugins,
	expectedPlugins map[string]tcp.Plugin,
	comparePointers bool,
) {

	for name, tcpPlugin := range expectedPlugins {
		assert.Contains(t, allPlugins.TCPPlugins(), name)
		if _, ok := allPlugins.TCPPlugins()[name]; !ok {
			t.Fail()
		}

		if comparePointers {
			assert.Equal(t, tcpPlugin, allPlugins.TCPPlugins()[name])
		}
	}
}

func newLogger() log_api.Logger {
	return log.New(true)
}

func TestPlugins(t *testing.T) {
	t.Run("HTTPPlugins", func(t *testing.T) {
		httpPlugins := mockPlugins().HTTPPlugins()

		assert.NotNil(t, httpPlugins)
		if httpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, httpPlugins, mockHTTPPlugins)
	})

	t.Run("TCPPlugins", func(t *testing.T) {
		tcpPlugins := mockPlugins().TCPPlugins()

		assert.NotNil(t, tcpPlugins)
		if tcpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, tcpPlugins, mockTCPPlugins)
	})
}

func TestCompatiblePluginVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", CompatiblePluginAPIVersion)
}

func TestAllAvailablePlugins(t *testing.T) {
	t.Run("Assembles internal and external plugins", func(t *testing.T) {
		allPlugins, err := AllAvailablePluginsWithOptions(
			"",
			"",
			mockGetInternalPlugins,
			mockGetExternalPlugins,
			newLogger(),
		)
		assert.Nil(t, err)
		if err != nil {
			t.Fail()
		}

		assert.Equal(t, 5, len(allPlugins.HTTPPlugins()))
		assert.Equal(t, 6, len(allPlugins.TCPPlugins()))

		assertMapContainsHTTPPlugins(t, allPlugins, mockInternalPlugins().HTTPPlugins(), false)
		assertMapContainsHTTPPlugins(t, allPlugins, mockExternalPlugins().HTTPPlugins(), false)

		assertMapContainsTCPPlugins(t, allPlugins, mockInternalPlugins().TCPPlugins(), false)
		assertMapContainsTCPPlugins(t, allPlugins, mockExternalPlugins().TCPPlugins(), false)
	})

	t.Run("External plugins override same-named internal plugins", func(t *testing.T) {
		defaultExternalPlugins, _ := mockGetExternalPlugins("", "", newLogger())

		httpPlugins := defaultExternalPlugins.HTTPPlugins()
		httpPlugins["intHTTP2"] = mockHTTPPlugin{}

		tcpPlugins := defaultExternalPlugins.TCPPlugins()
		tcpPlugins["intTCP1"] = mockTCPPlugin{}

		allExternalPlugins := Plugins{
			HTTPPluginsByID: httpPlugins,
			TCPPluginsByID:  tcpPlugins,
		}

		mockGetExternalPluginsWithOverride := func(
			pluginDir string,
			checksumfile string,
			logger log_api.Logger,
		) (AvailablePlugins, error) {
			return &allExternalPlugins, nil
		}

		allInternalPlugins, _ := mockGetInternalPlugins()
		mockGetInternalPluginsWithOverride := func() (AvailablePlugins, error) {
			return allInternalPlugins, nil
		}

		allPlugins, err := AllAvailablePluginsWithOptions(
			"",
			"",
			mockGetInternalPluginsWithOverride,
			mockGetExternalPluginsWithOverride,
			newLogger(),
		)

		assert.Nil(t, err)
		if err != nil {
			t.Fail()
		}

		assert.Equal(t, 5, len(allPlugins.HTTPPlugins()))
		assert.Equal(t, 6, len(allPlugins.TCPPlugins()))

		// Remove the overwritten plugins from checked equality maps
		delete(allInternalPlugins.HTTPPlugins(), "intHTTP2")
		delete(allInternalPlugins.TCPPlugins(), "intTCP1")

		assertMapContainsHTTPPlugins(t, allPlugins, allInternalPlugins.HTTPPlugins(), true)
		assertMapContainsHTTPPlugins(t, allPlugins, allExternalPlugins.HTTPPlugins(), true)

		assertMapContainsTCPPlugins(t, allPlugins, allInternalPlugins.TCPPlugins(), true)
		assertMapContainsTCPPlugins(t, allPlugins, allExternalPlugins.TCPPlugins(), true)
	})

	t.Run("Correct param info is passed to ExternalPluginLookupFunc", func(t *testing.T) {
		expectedDir := "fooDir"
		expectedChecksumfile := "fooChecksum"
		expectedLogger := newLogger()

		var mockGetExternalPluginsWithVerification = func(
			pluginDir string,
			checksumfile string,
			logger log_api.Logger,
		) (AvailablePlugins, error) {
			assert.Equal(t, expectedDir, pluginDir)
			assert.Equal(t, expectedChecksumfile, checksumfile)
			assert.Equal(t, expectedLogger, logger)
			return nil, errors.New("foo")
		}

		AllAvailablePluginsWithOptions(
			expectedDir,
			expectedChecksumfile,
			mockGetInternalPlugins,
			mockGetExternalPluginsWithVerification,
			expectedLogger,
		)
	})

	t.Run("InternalPluginLookupFunc errors are propagated", func(t *testing.T) {
		mockErr := errors.New("mock InternalPluginLookupFunc error")
		var mockGetInternalPluginsWithError = func() (AvailablePlugins, error) {
			return nil, mockErr
		}

		_, err := AllAvailablePluginsWithOptions(
			"directory",
			"hashfile",
			mockGetInternalPluginsWithError,
			mockGetExternalPlugins,
			newLogger(),
		)

		assert.Equal(t, mockErr, err)
	})

	t.Run("ExternalPluginLookupFunc errors are propagated", func(t *testing.T) {
		mockErr := errors.New("mock ExternalPluginLookupFunc error")
		var mockGetExternalPluginsWithError = func(
			pluginDir string,
			checksumfile string,
			logger log_api.Logger,
		) (AvailablePlugins, error) {
			return nil, mockErr
		}

		_, err := AllAvailablePluginsWithOptions(
			"directory",
			"hashfile",
			mockGetInternalPlugins,
			mockGetExternalPluginsWithError,
			newLogger(),
		)

		assert.Equal(t, mockErr, err)
	})
}
