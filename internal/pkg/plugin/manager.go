package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"plugin"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v1"

	"github.com/cyberark/secretless-broker/internal/app/secretless"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

var _SupportedFileSuffixes = []string{".so"}

func _IsDynamicLibrary(file os.FileInfo) bool {
	fileName := file.Name()
	for _, suffix := range _SupportedFileSuffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	return false
}

// Manager contains the main proxy and all connection managers, listener factories,
// and handler factories that are available
type Manager struct {
	configReloadMutex     sync.Mutex
	ConfigurationManagers map[string]plugin_v1.ConfigurationManager
	ConnectionManagers    map[string]plugin_v1.ConnectionManager
	DebugFlag             bool
	HandlerFactories      map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler
	ListenerFactories     map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener
	ProviderFactories     map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error)
	Proxy                 secretless.Proxy
}

var _singleton *Manager
var _once sync.Once

// GetManager returns the manager, and creates it if it doesn't exist
func GetManager() *Manager {
	_once.Do(func() {
		_singleton = &Manager{
			configReloadMutex:     sync.Mutex{},
			ConfigurationManagers: make(map[string]plugin_v1.ConfigurationManager),
			ConnectionManagers:    make(map[string]plugin_v1.ConnectionManager),
			DebugFlag:             false,
			HandlerFactories:      make(map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler),
			ListenerFactories:     make(map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener),
			ProviderFactories:     make(map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error)),
		}
	})

	return _singleton
}

// ConfigurationChanged is an interface adapter for plugin_v1.ConfigurationChangedHandler
func (manager *Manager) ConfigurationChanged(configManagerName string, newConfig config.Config) error {
	log.Printf("Configuration manager '%s' provided new configuration...",
		configManagerName)
	return manager._ReloadConfig(newConfig)
}

func (manager *Manager) _ReloadConfig(newConfig config.Config) error {
	log.Println("Reloading...")
	manager.configReloadMutex.Lock()

	log.Println("----------------------------")

	if manager.DebugFlag == true {
		configStr, _ := yaml.Marshal(newConfig)
		log.Printf("New configuration: %s", configStr)

		// Range uses a struct copy so we can't mod the actual items directly
		for index := range newConfig.Listeners {
			newConfig.Listeners[index].Debug = true
		}
		for index := range newConfig.Handlers {
			newConfig.Handlers[index].Debug = true
		}
	}

	manager.Proxy.Config = newConfig

	// TODO: Use a more robust and async way to detect when we can change the underlying config
	go func() {
		// Since our listeners/handlers might be still starting when the next reload even occurs
		// changing the underlying configuration is a really bad idea (that causes crashes) so
		// for now we add a small delay before processing the next reload.
		time.Sleep(1 * time.Second)
		manager.configReloadMutex.Unlock()
	}()

	return manager.Proxy.ReloadListeners()
}

func (manager *Manager) _RegisterShutdownSignalHandlers() {
	log.Println("Registering shutdown signal listeners...")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		exitSignal := <-signalChannel
		log.Printf("Intercepted exit signal '%v'. Cleaning up...", exitSignal)

		manager.Shutdown()

		log.Printf("Exiting...")
		os.Exit(0)
	}()
}

// RegisterSignalHandlers registers shutdown and reload signal handlers
func (manager *Manager) RegisterSignalHandlers() {
	manager._RegisterShutdownSignalHandlers()
}

// LoadConfigurationFile creates a configuration instance from a filesystem path
func (manager *Manager) LoadConfigurationFile(configFile string) config.Config {
	configuration, err := config.LoadFromFile(configFile)

	if err != nil {
		log.Fatal(err)
	}

	return configuration
}

// _IsSupportedPluginAPIVersion returns a boolean indicating if we support
// the interface API version that plugin is advertising
// TODO: Support a list of supported versions
func _IsSupportedPluginAPIVersion(version string) bool {
	return version == "0.0.8"
}

// LoadPlugin tries to load all internal plugin info strings
func _GetPluginInfo(pluginObj *plugin.Plugin) (map[string]string, error) {
	rawPluginInfo, err := pluginObj.Lookup("PluginInfo")
	if err != nil {
		return nil, err
	}
	pluginInfo := *rawPluginInfo.(*map[string]string)

	return pluginInfo, nil
}

// _LoadConfigurationManagers appends all configuration managers from the plugin to the pluginManager
func _LoadConfigurationManagers(manager *Manager, pluginObj *plugin.Plugin, pluginName string) error {
	rawConfigManagerPluginsFunc, err := pluginObj.Lookup("GetConfigurationManagers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	configManagerPluginsFunc, ok := rawConfigManagerPluginsFunc.(func() map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager)
	if !ok {
		return errors.New("ERROR: Could not cast GetConfigurationManagers to proper type")
	}
	configManagerPlugins := configManagerPluginsFunc()

	for managerID, configManagerPluginFactory := range configManagerPlugins {
		options := plugin_v1.ConfigurationManagerOptions{
			Name: managerID,
		}
		configManagerPlugin := configManagerPluginFactory(options)
		manager.ConfigurationManagers[managerID] = configManagerPlugin
		log.Printf("Configuration manager '%s' added from plugin %s", managerID, pluginName)
	}

	return nil
}

// _LoadConnectionManagers appends all managers from the plugin to the pluginManager
func _LoadConnectionManagers(manager *Manager, pluginObj *plugin.Plugin, pluginName string) error {

	rawManagerPluginsFunc, err := pluginObj.Lookup("GetConnectionManagers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	connManagerPluginsFunc, ok := rawManagerPluginsFunc.(func() map[string]func() plugin_v1.ConnectionManager)
	if !ok {
		return errors.New("ERROR: Could not cast GetConnectionManagers to proper type")
	}
	connManagerPlugins := connManagerPluginsFunc()

	for managerID, connManagerPluginFactory := range connManagerPlugins {
		connManagerPlugin := connManagerPluginFactory()
		manager.ConnectionManagers[managerID] = connManagerPlugin
		log.Printf("Connection manager '%s' added from plugin %s", managerID, pluginName)
	}

	return nil
}

// _LoadHandlersFromPlugin appends all handlers from the plugin to the pluginManager
func _LoadHandlersFromPlugin(manager *Manager, pluginObj *plugin.Plugin, pluginName string) error {
	rawHandlerPluginsFunc, err := pluginObj.Lookup("GetHandlers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	handlerPluginsFunc, ok := rawHandlerPluginsFunc.(func() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler)
	if !ok {
		return errors.New("ERROR: Could not cast GetcwHandlers to proper type")
	}
	handlerPlugins := handlerPluginsFunc()
	for handlerID, handlerFactory := range handlerPlugins {
		manager.HandlerFactories[handlerID] = handlerFactory
		log.Printf("Handler factory '%s' added from plugin %s", handlerID, pluginName)
	}

	return nil
}

// _LoadListenersFromPlugin appends all listeners from the plugin to the pluginManager
func _LoadListenersFromPlugin(manager *Manager, pluginObj *plugin.Plugin, pluginName string) error {
	rawListenerPluginsFunc, err := pluginObj.Lookup("GetListeners")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	listenerPluginsFunc, ok := rawListenerPluginsFunc.(func() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener)
	if !ok {
		return errors.New("ERROR: Could not cast GetListeners to proper type")
	}
	listenerPlugins := listenerPluginsFunc()
	for listenerID, listenerFactory := range listenerPlugins {
		manager.ListenerFactories[listenerID] = listenerFactory
		log.Printf("Listener factory '%s' added from plugin %s", listenerID, pluginName)
	}

	return nil
}

// _LoadProvidersFromPlugin appends all providers from the plugin to the pluginManager
func _LoadProvidersFromPlugin(manager *Manager, pluginObj *plugin.Plugin, pluginName string) error {
	rawProviderPluginsFunc, err := pluginObj.Lookup("GetProviders")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	providerPluginsFunc, ok := rawProviderPluginsFunc.(func() map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error))
	if !ok {
		return errors.New("ERROR: Could not cast GetProviders to proper type")
	}
	providerPlugins := providerPluginsFunc()
	for providerID, providerFactory := range providerPlugins {
		manager.ProviderFactories[providerID] = providerFactory
		log.Printf("Provider factory '%s' added from plugin %s", providerID, pluginName)
	}

	return nil
}

func (manager *Manager) _RunHandler(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler {
	// Ensure that we have this handler
	if _, ok := manager.HandlerFactories[id]; !ok {
		log.Panicf("Error! Unrecognized handler id '%s'", id)
	}

	handler := manager.HandlerFactories[id](options)
	manager.CreateHandler(handler, options.ClientConnection)
	return handler
}

func (manager *Manager) _RunListener(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener {
	// Ensure that we have this listener
	if _, ok := manager.ListenerFactories[id]; !ok {
		log.Panicf("Error: Unrecognized listener id '%s'", id)
	}

	return manager.ListenerFactories[id](options)
}

// Run is the main wait loop once we load all the plugins
func (manager *Manager) Run(configManagerID string, configManagerSpec string, debugSwitch bool) error {
	manager.DebugFlag = debugSwitch

	// We dont want any reloads happening until we are fully running
	manager.configReloadMutex.Lock()

	if len(configManagerID) > 0 {
		manager.InitializeConfigurationManager(configManagerID, configManagerSpec)
	}

	configuration := config.Config{}
	manager.InitializeConnectionManagers(configuration)

	log.Println("Initialization of plugins done.")

	log.Println("Initializing the proxy...")
	resolver := NewResolver(manager.ProviderFactories, manager, nil)

	manager.Proxy = secretless.Proxy{
		Config:          configuration,
		EventNotifier:   manager,
		Resolver:        resolver,
		RunHandlerFunc:  manager._RunHandler,
		RunListenerFunc: manager._RunListener,
	}

	manager.configReloadMutex.Unlock()

	manager.Proxy.Run()

	return nil
}

// LoadInternalPlugins loads all handlers/listeners/managers that are included in secretless by default
func (manager *Manager) LoadInternalPlugins() error {
	// Load all internal ConfigurationManagers
	for configManagerID, configManagerFactory := range secretless.InternalConfigurationManagers {
		options := plugin_v1.ConfigurationManagerOptions{Name: configManagerID}
		manager.ConfigurationManagers[configManagerID] = configManagerFactory(options)
	}
	configManagerNames := reflect.ValueOf(manager.ConfigurationManagers).MapKeys()
	log.Printf("- ConfigurationManagers: %v", configManagerNames)

	// Load all internal ConnectionManagers
	for connectionManagerID, connectionManagerFactory := range secretless.InternalConnectionManagers {
		manager.ConnectionManagers[connectionManagerID] = connectionManagerFactory()
	}
	connectionManagerNames := reflect.ValueOf(manager.ConnectionManagers).MapKeys()
	log.Printf("- ConnectionManagers: %v", connectionManagerNames)

	// Load all internal Providers
	for providerID, providerFactory := range secretless.InternalProviders {
		manager.ProviderFactories[providerID] = providerFactory
	}
	providerNames := reflect.ValueOf(manager.ProviderFactories).MapKeys()
	log.Printf("- Providers: %v", providerNames)

	// Load all internal Handlers
	for handlerID, handlerFactory := range secretless.InternalHandlers {
		manager.HandlerFactories[handlerID] = handlerFactory
	}
	handlerNames := reflect.ValueOf(manager.HandlerFactories).MapKeys()
	log.Printf("- Handlers: %v", handlerNames)

	// Load all internal Listeners
	for listenerID, listenerFactory := range secretless.InternalListeners {
		manager.ListenerFactories[listenerID] = listenerFactory
	}
	listenerNames := reflect.ValueOf(manager.ListenerFactories).MapKeys()
	log.Printf("- Listeners: %v", listenerNames)

	log.Println("Completed loading internal plugins.")
	return nil
}

// LoadLibraryPlugins loads all shared object files in `path`
func (manager *Manager) LoadLibraryPlugins(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !_IsDynamicLibrary(file) {
			continue
		}

		// Load shared library object
		pluginObj, err := plugin.Open(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		// Grab version of interface that the plugin is advertising
		rawPluginAPIVersion, err := pluginObj.Lookup("PluginAPIVersion")
		if err != nil {
			log.Println(err)
			continue
		}
		pluginAPIVersion := *rawPluginAPIVersion.(*string)

		// Bail if plugin interface API is an unsupported version
		log.Printf("Trying to load %s (API v%s)...", file.Name(), pluginAPIVersion)
		if !_IsSupportedPluginAPIVersion(pluginAPIVersion) {
			log.Printf("ERROR! Plugin %s is using an unsupported API version! Skipping!", file.Name())
			continue
		}

		// Get plugin info
		pluginInfo, err := _GetPluginInfo(pluginObj)
		if err != nil {
			continue
		}

		// Print out the fields in the info object
		formattedInfo, err := json.MarshalIndent(pluginInfo, "", "  ")
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s info:", file.Name())
		log.Println(string(formattedInfo))

		// Load configuration managers
		if err := _LoadConfigurationManagers(manager, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load connection managers
		if err := _LoadConnectionManagers(manager, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load providers
		if err := _LoadProvidersFromPlugin(manager, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load handlers
		if err := _LoadHandlersFromPlugin(manager, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load listeners
		if err := _LoadListenersFromPlugin(manager, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}
	}

	return nil
}

// InitializeConfigurationManager takes a configuration manager and initializes it
func (manager *Manager) InitializeConfigurationManager(id string, configSpec string) error {
	// Ensure that we have this configuration manager
	if _, ok := manager.ConfigurationManagers[id]; !ok {
		log.Panicf("Error! Unrecognized configuration manager id '%s'", id)
	}

	log.Printf("Initializing configuration manager '%s'...", id)
	configManagerPlugin := manager.ConfigurationManagers[id]

	err := configManagerPlugin.Initialize(manager, configSpec)
	if err != nil {
		log.Fatalf("Failed to initialize configuration manager '%s': %s\n",
			id,
			err.Error())
	}

	log.Printf("Configuration manager '%s' initialized (configSpec: '%s').",
		id, configSpec)

	return nil
}

// InitializeConnectionManagers loops through the connection managers and initializes them
func (manager *Manager) InitializeConnectionManagers(c config.Config) error {
	log.Println("Initializing connection managers...")
	for managerID, connManagerPlugin := range manager.ConnectionManagers {
		err := connManagerPlugin.Initialize(c, manager._ReloadConfig)
		if err != nil {
			log.Printf("Failed to initialize connection manager '%s': %s\n", managerID, err.Error())

			// TODO: Decide if manager initialization erroring is a fatal error.
			//       For now we treat it as a non-fatal error.
			delete(manager.ConnectionManagers, managerID)
			continue
		}
		log.Printf("Connection manager '%s' initialized.", managerID)
	}

	return nil
}

// CreateListener loops through the connection managers and creates the listener l
func (manager *Manager) CreateListener(l plugin_v1.Listener) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.CreateListener(l)
	}
}

// NewConnection loops through the connection managers and adds a connection to the listener l
func (manager *Manager) NewConnection(l plugin_v1.Listener, c net.Conn) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.NewConnection(l, c)
	}
}

// CloseConnection loops through the connection managers and closes the connection c
func (manager *Manager) CloseConnection(c net.Conn) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.CloseConnection(c)
	}
}

// CreateHandler loops through the connection managers to create the handler h
func (manager *Manager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.CreateHandler(h, c)
	}
}

// DestroyHandler loops through the connection managers to destroy the handler h
func (manager *Manager) DestroyHandler(h plugin_v1.Handler) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.DestroyHandler(h)
	}
}

// ResolveVariable loops through the connection managers to resolve the variable specified
func (manager *Manager) ResolveVariable(provider plugin_v1.Provider, id string, value []byte) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.ResolveVariable(provider, id, value)
	}
}

// ClientData loops through the connection managers to proxy data from the client
func (manager *Manager) ClientData(c net.Conn, buf []byte) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.ClientData(c, buf)
	}
}

// ServerData loops through the connection managers to proxy data from the server
func (manager *Manager) ServerData(c net.Conn, buf []byte) {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.ServerData(c, buf)
	}
}

// Shutdown calls Shutdown on the Proxy and all the connection managers, concurrently
func (manager *Manager) Shutdown() {
	manager.Proxy.Shutdown()

	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.Shutdown()
	}
}
