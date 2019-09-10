package configurationmanagers

import (
	"github.com/cyberark/secretless-broker/internal/configurationmanagers/configfile"
	"github.com/cyberark/secretless-broker/internal/configurationmanagers/kubernetes/crd"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// ConfigurationManagerFactories contains the list of built-in factories
var ConfigurationManagerFactories = map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager{
	configfile.PluginName: configfile.ConfigurationManagerFactory,
	crd.PluginName:        crd.ConfigurationManagerFactory,
}
