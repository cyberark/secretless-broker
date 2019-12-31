package configfile

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"syscall"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

const (
	// ConfigFileName is the default name expected from configuration files
	ConfigFileName = "secretless.yml"

	// PluginName is the external name that this plugin will be identified by
	PluginName = "configfile"
)

type configurationManager struct {
	ConfigChangedFunc func(string, config_v2.Config) error
	Name              string
}

// Local adapter struct that changes the ConfigurationChangedHandler events to
// chan config_v2.Config messages
type changeHandler struct {
	ConfigChangeChan chan config_v2.Config
}

func (ch *changeHandler) ConfigurationChanged(_ string, config config_v2.Config) error {
	ch.ConfigChangeChan <- config
	return nil
}

func (configManager *configurationManager) getConfigFilePreferenceOrder() ([]string, error) {
	configFileOrder := []string{
		"./" + ConfigFileName,
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	if len(currentUser.HomeDir) > 0 {
		homeDirConfigFile := "." + ConfigFileName
		fullHomeConfigFilePath := path.Join(currentUser.HomeDir, homeDirConfigFile)
		configFileOrder = append(configFileOrder, fullHomeConfigFilePath)
	}

	configFileOrder = append(configFileOrder,
		path.Join("/etc", ConfigFileName))

	return configFileOrder, nil
}

func (configManager *configurationManager) registerReloadSignalHandlers(configFile string,
	changeHandler plugin_v1.ConfigurationChangedHandler) {
	log.Println("Registering reload signal listeners...")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGUSR1)

	go func() {
		for {
			reloadSignal := <-signalChannel
			log.Printf("Intercepted reload signal '%v'. Reloading (from '%s')...",
				reloadSignal, configFile)

			configuration, err := config.LoadFromFile(configFile)
			if err != nil {
				log.Fatalf(err.Error())
			}
			changeHandler.ConfigurationChanged(configManager.Name, configuration)
		}
	}()
}

// registerConfigFileWatcher adds a configuration file change trigger for reloads
func (configManager *configurationManager) registerConfigFileWatcher(configFile string) {
	onChangeRunner := func() {
		log.Println("Sending reload signal (SIGUSR1)...")
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	}

	AttachWatcher(configFile, onChangeRunner)
}

func (configManager *configurationManager) onGoodConfigLoad(configuration config_v2.Config,
	changeHandler plugin_v1.ConfigurationChangedHandler, configFilePath string, watchFile bool) error {

	go func() {
		changeHandler.ConfigurationChanged(configManager.Name, configuration)
		if watchFile == true {
			configManager.registerConfigFileWatcher(configFilePath)
		}

		configManager.registerReloadSignalHandlers(configFilePath, changeHandler)
	}()

	return nil
}

// Initialize implements plugin_v1.ConfigurationManager
func (configManager *configurationManager) Initialize(changeHandler plugin_v1.ConfigurationChangedHandler,
	configSpecQuery string) error {

	configSpecObject, err := url.Parse(configSpecQuery)
	if err != nil {
		return err
	}

	configFilePath := configSpecObject.Path

	watchFile := false
	if watchFileParam, ok := configSpecObject.Query()["watch"]; ok == true {
		watchFileParamValue, err := strconv.ParseBool(watchFileParam[0])
		if err == nil {
			watchFile = watchFileParamValue
		}
	}

	if len(configFilePath) > 0 {
		log.Printf("Trying to load configuration file: %s", configFilePath)
		configuration, err := config.LoadFromFile(configFilePath)
		if err != nil {
			return err
		}

		return configManager.onGoodConfigLoad(configuration, changeHandler, configFilePath,
			watchFile)
	}

	configFileOrder, err := configManager.getConfigFilePreferenceOrder()
	if err != nil {
		return err
	}

	for _, configFilePath := range configFileOrder {
		log.Printf("Trying to load %s...", configFilePath)

		configuration, err := config.LoadFromFile(configFilePath)
		if err == nil {
			log.Printf("Configuration file %s loaded", configFilePath)
			return configManager.onGoodConfigLoad(configuration, changeHandler, configFilePath, watchFile)
		}

		log.Printf("WARN: Could not load %s: '%s'. Skipping...", configFilePath, err)

		continue
	}

	return fmt.Errorf("ERROR: Unable to locate any working configuration files")
}

// GetName implements plugin_v1.ConfigurationManager
func (configManager *configurationManager) GetName() string {
	return configManager.Name
}

// ConfigurationManagerFactory returns a file-based ConfigurationManager instance
func ConfigurationManagerFactory(options plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager {
	return &configurationManager{
		Name: options.Name,
	}
}

// NewConfigChannel returns a file-based ConfigurationManager channel object
func NewConfigChannel(configfile string, fsWatchEnabled bool) (<-chan config_v2.Config, error) {
	configChangedChan := make(chan config_v2.Config)

	manager := &configurationManager{}

	cfgChangeHandler := changeHandler{
		ConfigChangeChan: configChangedChan,
	}

	// Managers expect parameters to be passed in as URL params
	cfgSpec := fmt.Sprintf("%s?watch=%t", configfile, fsWatchEnabled)

	err := manager.Initialize(&cfgChangeHandler, cfgSpec)
	if err != nil {
		return nil, err
	}

	return cfgChangeHandler.ConfigChangeChan, nil
}
