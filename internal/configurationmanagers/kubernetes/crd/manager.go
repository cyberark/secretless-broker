package crd

import (
	"log"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

type configurationManager struct {
	ConfigChangedChan chan config_v2.Config
	FilterSpec        string
	Name              string
}

// NewConfigChannel returns a CRD-based ConfigurationManager channel object
// TODO: Also return the name of configuration provider module
func NewConfigChannel(configSpec string) (<-chan config_v2.Config, error) {
	configManager := &configurationManager{
		ConfigChangedChan: make(chan config_v2.Config),
	}

	// Inject CRD just in case we don't have it in the cluster yet
	if err := InjectCRD(); err != nil {
		return nil, err
	}

	// TODO: Wait for CRD to be online

	// Watch for changes
	err := RegisterCRDListener(meta_v1.NamespaceAll, configSpec, configManager)
	if err != nil {
		return nil, err
	}

	return configManager.ConfigChangedChan, nil
}

// CRDAdded implements crd.ResourceEventHandler
func (configManager *configurationManager) CRDAdded(crdConfiguration *api_v1.Configuration) {
	newConfig, err := config.LoadFromCRD(*crdConfiguration)
	if err != nil {
		log.Printf("%s: WARN: New CRD could not be turned into a config.Config!", PluginName)
		return
	}

	configManager.ConfigChangedChan <- newConfig
}

// CRDDeleted implements crd.ResourceEventHandler
func (configManager *configurationManager) CRDDeleted(crdConfiguration *api_v1.Configuration) {
	log.Printf("%s: WARN: CRDDeleted - setting empty config!", PluginName)

	// TODO: Do something of value here
	newConfig := config_v2.Config{}

	configManager.ConfigChangedChan <- newConfig
}

// CRDUpdated implements crd.ResourceEventHandler
func (configManager *configurationManager) CRDUpdated(oldCRDConfiguration *api_v1.Configuration,
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

	configManager.ConfigChangedChan <- newConfig
}
