package sharedobj

import (
	"errors"
	"fmt"
	"testing"

	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	loggerMock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj/mock"
	"github.com/stretchr/testify/assert"
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

	// Test detection of external plugin IDs conflicting with internal plugins.
	TestCases := []struct {
		description   string
		addPluginType string // "HTTP" or "TCP"
		addPluginID   string
		expectPanic   bool
	}{
		{
			description: "There are no panics without duplicate plugin names",
		},
		{
			description:   "Panic when an external HTTP plugin and internal HTTP plugin use the same ID",
			addPluginType: "HTTP",
			addPluginID:   "intHTTP1",
			expectPanic:   true,
		},
		{
			description:   "Panic when an external HTTP plugin and internal TCP plugin use the same ID",
			addPluginType: "HTTP",
			addPluginID:   "intTCP2",
			expectPanic:   true,
		},
		{
			description:   "Panic when an external TCP plugin and internal HTTP plugin use the same ID",
			addPluginType: "TCP",
			addPluginID:   "intHTTP2",
			expectPanic:   true,
		},
		{
			description:   "Panic when an external TCP plugin and an internal TCP plugin use the same ID",
			addPluginType: "TCP",
			addPluginID:   "intTCP1",
			expectPanic:   true,
		},
	}
	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			// Set up baseline mock plugins
			extHTTPPlugins := mock.ExternalPlugins().HTTPPlugins()
			allHTTPPlugins := mock.AllHTTPPlugins()
			extTCPPlugins := mock.ExternalPlugins().TCPPlugins()
			allTCPPlugins := mock.AllTCPPlugins()

			// Add an additional mock plugin for this test case
			id := tc.addPluginID
			switch tc.addPluginType {
			case "HTTP":
				plugin := mock.NewHTTPPlugin("conflicting name")
				extHTTPPlugins[id] = plugin
				allHTTPPlugins[id] = plugin
			case "TCP":
				plugin := mock.NewTCPPlugin("conflicting name")
				extTCPPlugins[id] = plugin
				allTCPPlugins[id] = plugin
			}

			// Create an external plugins lookup func
			allExternalPlugins := Plugins{
				HTTPPluginsByID: extHTTPPlugins,
				TCPPluginsByID:  extTCPPlugins,
			}
			externalPlugins := func(
				pluginDir string,
				checksumfile string,
				logger log_api.Logger,
			) (plugin.AvailablePlugins, error) {
				return &allExternalPlugins, nil
			}

			// Run the test subject
			mockLogger := loggerMock.NewLogger()
			allPlugins, err := AllAvailablePluginsWithOptions(
				"",
				"",
				mock.GetInternalPlugins,
				externalPlugins,
				mockLogger,
			)

			// Check test results
			if tc.expectPanic {
				expectedPanic := fmt.Sprintf(
					pluginConflictMessage,
					tc.addPluginType,
					tc.addPluginID)
				panic := mockLogger.Panics[0]
				assert.Contains(t, panic, expectedPanic)
			} else {
				assert.NoError(t, err)
				if err != nil {
					return
				}

				// Confirm expected HTTP plugins have been discovered
				assert.Equal(t, allHTTPPlugins, allPlugins.HTTPPlugins())

				// Confirm expected TCP plugins have been discovered
				assert.Equal(t, allTCPPlugins, allPlugins.TCPPlugins())
			}
		})
	}

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
