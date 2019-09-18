package mock

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

func InternalPlugins() plugin.AvailablePlugins {
	return &plugin.Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"intHTTP1": &MockHTTPPlugin{},
			"intHTTP2": &MockHTTPPlugin{},
			"intHTTP3": &MockHTTPPlugin{},
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"intTCP1": &MockTCPPlugin{},
			"intTCP2": &MockTCPPlugin{},
			"intTCP3": &MockTCPPlugin{},
		},
	}
}
