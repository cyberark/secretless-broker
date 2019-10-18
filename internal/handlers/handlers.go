package handlers

import (
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// HandlerFactories contains the list of built-in handler factories
var HandlerFactories = map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{}
