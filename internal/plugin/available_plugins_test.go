package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

type mockHTTPPlugin struct{ http.Plugin }
type mockTCPPlugin struct{ tcp.Plugin }

// These need to sit outside of the func since we are comparing them in the
// tests
var mockHTTPPlugins = map[string]http.Plugin{
	"one": mockHTTPPlugin{},
	"two": mockHTTPPlugin{},
}

var mockTCPPlugins = map[string]tcp.Plugin{
	"one":   mockTCPPlugin{},
	"two":   mockTCPPlugin{},
	"three": mockTCPPlugin{},
}

func getMockPlugins() AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: mockHTTPPlugins,
		TCPPluginsByID:  mockTCPPlugins,
	}
}

func TestPlugins(t *testing.T) {
	t.Run("HTTPPlugins", func(t *testing.T) {
		httpPlugins := getMockPlugins().HTTPPlugins()

		assert.NotNil(t, httpPlugins)
		if httpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, httpPlugins, mockHTTPPlugins)
	})

	t.Run("TCPPlugins", func(t *testing.T) {
		tcpPlugins := getMockPlugins().TCPPlugins()

		assert.NotNil(t, tcpPlugins)
		if tcpPlugins == nil {
			t.Fail()
		}

		assert.Equal(t, tcpPlugins, mockTCPPlugins)
	})
}
