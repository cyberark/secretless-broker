package listeners

import (
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// ListenerFactories contains the list of built-in listener factories
var ListenerFactories = map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{}
