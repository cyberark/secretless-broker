package sharedobj

import (
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/aws"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/basicauth"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/conjur"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/generic"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/cassandra"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// InternalPluginLookupFunc returns all available buiilt-in plugins.
type InternalPluginLookupFunc func() (plugin.AvailablePlugins, error)

// GetInternalPluginsFunc returns currently available built-in plugins.
func GetInternalPluginsFunc() (plugin.AvailablePlugins, error) {
	// New connectors should have an entry in the map below, according to their type (HTTP/TCP)
	return &Plugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"aws":          aws.GetHTTPPlugin(),
			"basic_auth":   basicauth.GetHTTPPlugin(),
			"conjur":       conjur.GetHTTPPlugin(),
			"generic_http": generic.GetHTTPPlugin(),
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"pg":    pg.GetTCPPlugin(),
			"mysql": mysql.GetTCPPlugin(),
			"mssql": mssql.GetTCPPlugin(),
			"cassandra": cassandra.GetTCPPlugin(),
		},
	}, nil
}

// InternalPlugins is used to enumerate built-in plugins.
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
