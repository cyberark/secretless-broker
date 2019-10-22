package sharedobj

import (
	"github.com/cyberark/secretless-broker/internal/proxyservice/http/aws"
	"github.com/cyberark/secretless-broker/internal/proxyservice/http/basicauth"
	"github.com/cyberark/secretless-broker/internal/proxyservice/http/conjur"
	"github.com/cyberark/secretless-broker/internal/proxyservice/tcp/mssql"
	"github.com/cyberark/secretless-broker/internal/proxyservice/tcp/mysql"
	"github.com/cyberark/secretless-broker/internal/proxyservice/tcp/pg"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// InternalPluginLookupFunc returns all available internal plugins.
type InternalPluginLookupFunc func() (plugin.AvailablePlugins, error)

// GetInternalPluginsFunc returns currently available internal plugins
// but for now, this list is empty since we have none implemented.
func GetInternalPluginsFunc() (plugin.AvailablePlugins, error) {
	return &Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"aws":        aws.GetHTTPPlugin(),
			"basic_auth": basicauth.GetHTTPPlugin(),
			"conjur":     conjur.GetHTTPPlugin(),
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"pg":    pg.GetTCPPlugin(),
			"mysql": mysql.GetTCPPlugin(),
			"mssql": mssql.GetTCPPlugin(),
		},
	}, nil
}

// InternalPlugins is used to enumerate internally-available plugins to the clients
// of this method.
func InternalPlugins(lookupFunc InternalPluginLookupFunc) (plugin.AvailablePlugins, error) {
	plugins, err := lookupFunc()
	if err != nil {
		return nil, err
	}

	if plugins == nil {
		plugins = &Plugins{
			HTTPPluginsByID: nil,
			TCPPluginsByID:  nil,
		}
	}

	return plugins, nil
}
