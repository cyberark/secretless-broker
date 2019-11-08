package v2_test

import (
	"testing"

	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	loggermock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj/mock"
	"github.com/stretchr/testify/assert"
)

// TODO: move the log mock public

func TestConfigEnv(t *testing.T) {

	// A dummy to fulfill the dependency.  Asserting on the Infof call isn't
	// really worthwhile.
	logger := loggermock.NewLogger()

	// Shared dependency for all the tests
	availPlugins := &mock.Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"HTTP1": mock.NewHTTPPlugin("HTTP1"),
			"HTTP2": mock.NewHTTPPlugin("HTTP2"),
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"TCP1": mock.NewTCPPlugin("TCP1"),
			"TCP2": mock.NewTCPPlugin("TCP2"),
		},
	}

	configEnv := v2.NewConfigEnv(logger, availPlugins)

	t.Run("succeeds when all connectors are available", func(t *testing.T) {
		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1"},
				{Name: "HTTP2 Service", Connector: "HTTP2"},
				{Name: "TCP1 Service", Connector: "TCP1"},
				{Name: "TCP2 Service", Connector: "TCP2"},
			},
		}

		err := configEnv.Prepare(cfg)
		assert.Nil(t, err)
	})

	t.Run("errors when connectors aren't available", func(t *testing.T) {
		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1"},
				{Name: "HTTP2 Service", Connector: "TCP2"},
				{Name: "Fake Service 1", Connector: "FAKE PLUGIN 1"},
				{Name: "Fake Service 2", Connector: "FAKE PLUGIN 2"},
			},
		}

		err := configEnv.Prepare(cfg)
		assert.NotNil(t, err)
		if err == nil {
			return
		}

		// Ensure error message mentions which plugins are missing.
		assert.Regexp(t, "FAKE PLUGIN 1", err)
		assert.Regexp(t, "FAKE PLUGIN 2", err)
	})

	// returns error when socket can't be delted
}
