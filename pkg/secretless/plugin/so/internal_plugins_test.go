package so

import (
	"errors"
	"testing"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

func getMockHTTPPlugins() plugin.AvailablePlugins {
	var mockInternalHTTPPlugins = map[string]http.Plugin{
		"one": mockHTTPPlugin{},
		"two": mockHTTPPlugin{},
	}

	var mockInternalTCPPlugins = map[string]tcp.Plugin{
		"one":   mockTCPPlugin{},
		"two":   mockTCPPlugin{},
		"three": mockTCPPlugin{},
	}

	return &Plugins{
		HTTPPluginsByID: mockInternalHTTPPlugins,
		TCPPluginsByID:  mockInternalTCPPlugins,
	}
}

func TestInternalPlugins(t *testing.T) {
	t.Run("InternalPluginFunc plugins are passed through", func(t *testing.T) {
		plugins, _ := InternalPlugins(func() (plugin.AvailablePlugins, error) {
			return getMockHTTPPlugins(), nil
		})

		assert.NotNil(t, plugins)
		if plugins == nil {
			t.Fail()
		}

		assert.Equal(t, plugins.HTTPPlugins(), getMockHTTPPlugins().HTTPPlugins())
		assert.Equal(t, plugins.TCPPlugins(), getMockHTTPPlugins().TCPPlugins())
	})

	t.Run("InternalPluginFunc does not pass nil plugins", func(t *testing.T) {
		plugins, _ := InternalPlugins(func() (plugin.AvailablePlugins, error) {
			return nil, nil
		})

		assert.NotNil(t, plugins)
		if plugins == nil {
			t.Fail()
		}

		assert.Equal(t, len(plugins.HTTPPlugins()), 0)
		assert.Equal(t, len(plugins.TCPPlugins()), 0)
	})

	t.Run("InternalPluginFunc errors are passed through", func(t *testing.T) {
		mockError := errors.New("Some error")
		plugins, err := InternalPlugins(func() (plugin.AvailablePlugins, error) {
			return nil, mockError
		})

		assert.Nil(t, plugins)
		assert.Error(t, err)
		assert.Equal(t, mockError, err)
	})
}

func TestGetInternalPlugins(t *testing.T) {
	t.Run("GetInternalPluginsFunc does not error out", func(t *testing.T) {
		_, err := GetInternalPluginsFunc()
		assert.Nil(t, err)
	})

	t.Run("GetInternalPluginsFunc returns the expected plugin list", func(t *testing.T) {
		internalPlugins, err := GetInternalPluginsFunc()
		assert.Nil(t, err)

		if err != nil {
			t.Fail()
		}

		assert.NotNil(t, internalPlugins.HTTPPlugins())
		assert.NotNil(t, internalPlugins.TCPPlugins())

		if internalPlugins.HTTPPlugins() != nil {
			assert.Equal(t, 0, len(internalPlugins.HTTPPlugins()))
		}

		if internalPlugins.TCPPlugins() != nil {
			assert.Equal(t, 0, len(internalPlugins.TCPPlugins()))
		}
	})
}
