package crd

import (
	"log"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

type configurationManager struct {
	ConfigChangedFunc func(string, config.Config) error
	FilterSpec        string
	Name              string
}

// Initialize implements plugin_v1.ConfigurationManager
func (configManager *ConfigurationManager) Initialize(changeHandler plugin_v1.ConfigurationChangedHandler,
	configSpec string) error {

	configManager.ConfigChangedFunc = changeHandler.ConfigurationChanged

	// Inject CRD just in case we don't have it in the cluster yet
	if err := InjectCRD(); err != nil {
		return err
	}

	// TODO: Wait for CRD to be online

	// Watch for changes
	return RegisterCRDListener(meta_v1.NamespaceAll, configSpec, configManager)
}

// CRDAdded implements crd.ResourceEventHandler
func (configManager *ConfigurationManager) CRDAdded(crdConfiguration *api_v1.Configuration) {
	newConfig, err := config.LoadFromCRD(*crdConfiguration)
	if err != nil {
		log.Printf("%s: WARN: New CRD could not be turned into a config.Config!", PluginName)
		return
	}

	configManager.ConfigChangedFunc(configManager.Name, newConfig)
}

// CRDDeleted implements crd.ResourceEventHandler
func (configManager *ConfigurationManager) CRDDeleted(crdConfiguration *api_v1.Configuration) {
	log.Printf("%s: WARN: CRDDeleted - setting empty config!", PluginName)

	// TODO: Do something of value here
	newConfig := config.Config{}
	configManager.ConfigChangedFunc(configManager.Name, newConfig)
}

// CRDUpdated implements crd.ResourceEventHandler
func (configManager *ConfigurationManager) CRDUpdated(oldCRDConfiguration *api_v1.Configuration,
	newCRDConfiguration *api_v1.Configuration) {

	oldConfig, err := config.LoadFromCRD(*oldCRDConfiguration)
	if err != nil {
		log.Printf("%s: WARN: Pre-update CRD could not be turned into a config.Config! %v",
			PluginName, err)
		return
	}

	newConfig, err := config.LoadFromCRD(*newCRDConfiguration)
	if err != nil {
		log.Printf("%s: WARN: Updated CRD could not be turned into a config.Config! %v",
			PluginName, err)
		return
	}

	// If the old config and the new config are the same, that means that we are getting our
	// periodic polling list back and that we shouldn't worry about it.
	if oldConfig.String() == newConfig.String() {
		return
	}

	configManager.ConfigChangedFunc(configManager.Name, newConfig)
}

// GetName implements plugin_v1.ConfigurationManager
func (configManager *ConfigurationManager) GetName() string {
	return configManager.Name
}

// ConfigurationManagerFactory returns a CRD ConfigurationManager instance
func ConfigurationManagerFactory(options plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager {
	return &configurationManager{
		Name: options.Name,
	}
}
