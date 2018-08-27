package configurationmanagers

import (
	"github.com/cyberark/secretless-broker/internal/app/secretless/configurationmanagers/configfile"
	"github.com/cyberark/secretless-broker/internal/app/secretless/configurationmanagers/kubernetes/crd"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// ConfigurationManagerFactories contains the list of built-in factories
var ConfigurationManagerFactories = map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager{
	configfile.PluginName: configfile.ConfigurationManagerFactory,
	crd.PluginName:        crd.ConfigurationManagerFactory,
}
