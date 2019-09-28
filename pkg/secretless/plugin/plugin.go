package plugin

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// AvailablePlugins is an interface that provides a list of all the available
// plugins for each type that the broker supports.
type AvailablePlugins interface {
	HTTPPlugins() map[string]http.Plugin
	TCPPlugins() map[string]tcp.Plugin
}

