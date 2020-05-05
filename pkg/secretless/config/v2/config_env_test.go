package v2_test

import (
	"fmt"
	"os"
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

	// Shared mocks and doubles

	logger := loggermock.NewLogger()

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

	// Mock to simulate file existing.
	//
	// NOTE: We're cheating a little bit by returning nil, since we're
	// relying on knowing that our implementation doesn't actually use
	// FileInfo.  That said, this should be ok since the test will likely
	// break cleanly if the implementation is changed to use FileInfo.
	getInfoForExistingFile := func(_ string) (os.FileInfo, error) {
		return nil, nil
	}

	// Tests

	t.Run("succeeds when all connectors are available", func(t *testing.T) {
		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1"},
				{Name: "HTTP2 Service", Connector: "HTTP2"},
				{Name: "TCP1 Service", Connector: "TCP1"},
				{Name: "TCP2 Service", Connector: "TCP2"},
			},
		}

		configEnv := v2.NewConfigEnv(logger, availPlugins)
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

		configEnv := v2.NewConfigEnv(logger, availPlugins)
		err := configEnv.Prepare(cfg)
		assert.NotNil(t, err)
		if err == nil {
			return
		}

		// Ensure error message mentions which plugins are missing.
		assert.Regexp(t, "FAKE PLUGIN 1", err)
		assert.Regexp(t, "FAKE PLUGIN 2", err)
	})

	// When a listenOn unix socket exists, it tries to delete it, and if the
	// deletion succeeds, it does not return an error.
	t.Run("can delete existing unix socket without error", func(t *testing.T) {

		// Mock to simulate successful file deletion.
		socketDeleted := ""
		deleteFile := func(name string) error {
			socketDeleted = name
			return nil
		}

		configEnv := v2.NewConfigEnvWithOptions(
			logger,
			availPlugins,
			getInfoForExistingFile,
			deleteFile,
		)

		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1", ListenOn: "unix://xxx"},
			},
		}

		err := configEnv.Prepare(cfg)
		assert.Nil(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, "xxx", socketDeleted)
	})

	t.Run("returns error if socket deletion fails", func(t *testing.T) {

		// Double to simulate failing file deletion.
		deleteFile := func(name string) error {
			return fmt.Errorf("failed to delete file")
		}

		configEnv := v2.NewConfigEnvWithOptions(
			logger,
			availPlugins,
			getInfoForExistingFile,
			deleteFile,
		)

		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1", ListenOn: "unix://xxx"},
			},
		}

		err := configEnv.Prepare(cfg)
		assert.NotNil(t, err)
		if err == nil {
			return
		}

		assert.Regexp(t, "delete", err)
	})

	t.Run("doesn't delete non-unix sockets", func(t *testing.T) {

		// Mock to simulate successful file deletion.
		deleteFileCalled := false
		deleteFile := func(name string) error {
			deleteFileCalled = true
			return nil
		}

		configEnv := v2.NewConfigEnvWithOptions(
			logger,
			availPlugins,
			getInfoForExistingFile,
			deleteFile,
		)

		cfg := v2.Config{
			Services: []*v2.Service{
				{Name: "HTTP1 Service", Connector: "HTTP1", ListenOn: "tcp://xxx"},
			},
		}

		err := configEnv.Prepare(cfg)
		assert.Nil(t, err)
		if err != nil {
			return
		}

		assert.False(t, deleteFileCalled, "deleteFile should not be called")
	})
}
