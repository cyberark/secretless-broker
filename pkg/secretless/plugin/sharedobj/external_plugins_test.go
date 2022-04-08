package sharedobj

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	loggerMock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj/mock"
)

var metadataTestCases = []struct {
	description   string
	rawPlugin     rawPlugin
	rawPluginName string
	expectedType  string
	expectedID    string
	expectedError string
}{
	{
		description:   "happy path",
		rawPlugin:     mock.RawPlugins["http1"],
		rawPluginName: "mock_plugin",
		expectedType:  "connector.http",
		expectedID:    "extHTTP1",
	},
	{
		description: "empty plugin id results in error",
		rawPlugin: mock.RawPlugin{
			PluginAPIVersion: "0.1.0",
			PluginType:       "connector.http",
		},
		rawPluginName: "mock_plugin",
		expectedError: "PluginInfo['id'] is blank",
	},
	{
		description: "empty plugin type results in error",
		rawPlugin: mock.RawPlugin{
			PluginAPIVersion: "0.1.0",
			PluginID:         "extHTTP1",
		},
		rawPluginName: "mock_plugin",
		expectedError: "PluginInfo['type'] is blank",
	},
	{
		description: "empty plugin api version in error",
		rawPlugin: mock.RawPlugin{
			PluginType: "connector.http",
			PluginID:   "extHTTP1",
		},
		rawPluginName: "mock_plugin",
		expectedError: "PluginInfo['pluginAPIVersion'] is blank",
	},
	{
		description: "unsupported plugin api version in error",
		rawPlugin: mock.RawPlugin{
			PluginType:       "connector.http",
			PluginID:         "extHTTP1",
			PluginAPIVersion: "0.2.0",
		},
		rawPluginName: "mock_plugin",
		expectedError: "plugin 'mock_plugin' (API v0.2.0) is not a supported API version (v0.1.0)",
	},
	{
		description: "plugin without PluginInfo results in error",
		rawPlugin: mock.InvalidPlugin{
			PluginAPIVersion: "0.1.0",
			PluginID:         "extHTTP1",
			PluginType:       "connector.tcp",
			ErrorOnLookup:    true,
		},
		rawPluginName: "mock_plugin",
		expectedError: "error on lookup",
	},
	{
		description: "plugin with wrong type for PluginInfo results in error",
		rawPlugin: mock.InvalidPlugin{
			PluginAPIVersion: "0.1.0",
			PluginID:         "extHTTP1",
			PluginType:       "connector.tcp",
		},
		rawPluginName: "mock_plugin",
		expectedError: "could not cast PluginInfo to proper type",
	},
}

func TestParsePluginMetadata(t *testing.T) {
	for _, tc := range metadataTestCases {
		t.Run(tc.description, func(t *testing.T) {
			// Run test
			pluginType, pluginID, err := parsePluginMetadata(tc.rawPlugin, tc.rawPluginName)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedType, pluginType)
				assert.Equal(t, tc.expectedID, pluginID)
			}
		})
	}
}

// directoryPluginLookup implements an external DirectoryPluginLookupFunc
// for testing.
func directoryPluginLookup(mockPlugins map[string]mock.RawPlugin) DirectoryPluginLookupFunc {
	return func(
		pluginDir string,
		checksumfile string,
		logger log.Logger,
	) (map[string]rawPlugin, error) {
		rawPlugins := map[string]rawPlugin{}
		for name, plugin := range mockPlugins {
			rawPlugins[name] = plugin
		}
		return rawPlugins, nil
	}
}

func TestExternalPlugins(t *testing.T) {
	t.Run("Assembles external plugins", func(t *testing.T) {
		externalPlugins, err := ExternalPluginsWithOptions(
			"",
			"",
			directoryPluginLookup(mock.RawPlugins),
			mock.NewLogger(),
		)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		expExternalPlugins := mock.ExternalPlugins()
		assert.EqualValues(t, expExternalPlugins.HTTPPlugins(), externalPlugins.HTTPPlugins())
		assert.EqualValues(t, expExternalPlugins.TCPPlugins(), externalPlugins.TCPPlugins())
	})

	// Test detection of conflicting external plugin names.
	testCases := []struct {
		description   string
		addPluginName string
		addPluginType string // "connector.http" or "connector.tcp"
		addPluginID   string
		expectPanic   bool
	}{
		{
			description: "There are no panics without duplicate plugin names",
		},
		{
			description:   "Two external HTTP plugins use same plugin ID",
			addPluginName: "new_http_plugin",
			addPluginType: "connector.http",
			addPluginID:   "extHTTP1",
			expectPanic:   true,
		},
		{
			description:   "Two external TCP plugins use same plugin ID",
			addPluginName: "new_tcp_plugin",
			addPluginType: "connector.tcp",
			addPluginID:   "extTCP2",
			expectPanic:   true,
		},
		{
			description:   "An HTTP and a TCP external plugin use same plugin ID",
			addPluginName: "new_tcp_plugin",
			addPluginType: "connector.tcp",
			addPluginID:   "extHTTP2",
			expectPanic:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Set up baseline test plugins
			testPlugins := map[string]mock.RawPlugin{}
			for name, rawPlugin := range mock.RawPlugins {
				testPlugins[name] = rawPlugin
			}
			expHTTPPlugins := mock.HTTPExternalPluginsByID()
			expTCPPlugins := mock.TCPExternalPluginsByID()

			// Add external plugin for this test case, if necessary
			if tc.addPluginName != "" {
				testPlugins[tc.addPluginName] = mock.RawPlugin{
					PluginAPIVersion: "0.1.0",
					PluginType:       tc.addPluginType,
					PluginID:         tc.addPluginID,
				}
				switch tc.addPluginType {
				case "connector.http":
					expHTTPPlugins[tc.addPluginName] = mock.NewHTTPPlugin(tc.addPluginID)
				case "connector.tcp":
					expTCPPlugins[tc.addPluginName] = mock.NewTCPPlugin(tc.addPluginID)
				}
			}

			// Run the test subject
			mockLogger := loggerMock.NewLogger()
			availablePlugins, err := ExternalPluginsWithOptions(
				"",
				"",
				directoryPluginLookup(testPlugins),
				mockLogger,
			)

			// Check test results
			if tc.expectPanic {
				expectedPanic := "conflicts with external plugin"
				panic := mockLogger.Panics[0]
				assert.Contains(t, panic, expectedPanic)
			} else {
				assert.NoError(t, err)
				if err != nil {
					return
				}
				// Confirm expected HTTP plugins have been discovered
				assert.Equal(t, expHTTPPlugins, availablePlugins.HTTPPlugins())

				// Confirm expected TCP plugins have been discovered
				assert.Equal(t, expTCPPlugins, availablePlugins.TCPPlugins())
			}
		})
	}
}
