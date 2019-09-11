package secretless

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

type Secretless interface {
	Start()
	Stop()
}

type AvailablePlugins interface {
	HTTPPlugins() map[string]http.Plugin
	TCPPlugins() map[string]tcp.Plugin
}

// stubs to be replaced when PRs for these arrive

//func AllAvailablePlugins(pluginDir string) (AvailablePlugins, error) {
//	return nil, nil
//}
