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

	"github.com/conjurinc/secretless/internal/app/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	"github.com/conjurinc/secretless/internal/pkg/global"
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
	ConnectionManagers map[string]plugin_v1.ConnectionManager
	HandlerFactories   map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler
	ListenerFactories  map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener
	ProviderFactories  map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider
	Proxy              secretless.Proxy
}

var _singleton *Manager
var _once sync.Once

// GetManager returns the manager, and creates it if it doesn't exist
func GetManager() *Manager {
	_once.Do(func() {
		_singleton = &Manager{
			ConnectionManagers: make(map[string]plugin_v1.ConnectionManager),
			HandlerFactories:   make(map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler),
			ListenerFactories:  make(map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener),
			ProviderFactories:  make(map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider),
		}
	})

	return _singleton
}

func (manager *Manager) _ReloadConfig(newConfig config.Config) error {
	log.Println("----------------------------")
	log.Println("Reloading...")
	manager.Proxy.Config = newConfig
	return manager.Proxy.ReloadListeners()
}

func (manager *Manager) _RegisterShutdownSignalHandlers() {
	log.Println("Registering shutdown signal listeners...")

	shutdownCh, cleanUpShutdownCh := global.ShutdownChCreator(syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		global.WaitForGlobalCleanUp()

		log.Printf("Exiting...")
		os.Exit(0)
	}()
	go func() {
		defer cleanUpShutdownCh()
		exitSignal := <-shutdownCh
		log.Printf("Intercepted exit signal '%v'. Cleaning up...", exitSignal)

		manager.Shutdown()
	}()
}

func (manager *Manager) _RegisterReloadSignalHandlers() {
	log.Println("Registering reload signal listeners...")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGUSR1)

	go func() {
		for {
			reloadSignal := <-signalChannel
			log.Printf("Intercepted reload signal '%v'. Reloading...", reloadSignal)
			manager._ReloadConfig(manager.Proxy.Config)
		}
	}()
}

// RegisterSignalHandlers registers shutdown and reload signal handlers
func (manager *Manager) RegisterSignalHandlers() {
	manager._RegisterShutdownSignalHandlers()
	manager._RegisterReloadSignalHandlers()
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

// _LoadConnectionManagers appends all managers from the plugin to the pluginManager
func _LoadConnectionManagers(manager *Manager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

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
		log.Printf("Manager '%s' added from plugin %s", managerID, pluginName)
	}

	return nil
}

// _LoadHandlersFromPlugin appends all handlers from the plugin to the pluginManager
func _LoadHandlersFromPlugin(manager *Manager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

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
func _LoadListenersFromPlugin(manager *Manager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

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
func _LoadProvidersFromPlugin(manager *Manager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	rawProviderPluginsFunc, err := pluginObj.Lookup("GetProviders")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	providerPluginsFunc, ok := rawProviderPluginsFunc.(func() map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider)
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

	return manager.HandlerFactories[id](options)
}

func (manager *Manager) _RunListener(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener {
	// Ensure that we have this listener
	if _, ok := manager.ListenerFactories[id]; !ok {
		log.Panicf("Error: Unrecognized listener id '%s'", id)
	}

	return manager.ListenerFactories[id](options)
}

// Run is the main wait loop once we load all the plugins
func (manager *Manager) Run(configuration config.Config) error {
	manager.InitializeConnectionManagers(configuration)

	resolver := &Resolver{
		EventNotifier:     manager,
		ProviderFactories: manager.ProviderFactories,
	}

	manager.Proxy = secretless.Proxy{
		Config:          configuration,
		EventNotifier:   manager,
		Resolver:        resolver,
		RunHandlerFunc:  manager._RunHandler,
		RunListenerFunc: manager._RunListener,
	}

	manager.Proxy.Run()

	return nil
}

// LoadInternalPlugins loads all handlers/listeners/managers that are included in secretless by default
func (manager *Manager) LoadInternalPlugins(config config.Config) error {
	// Load all internal ConnectionManagers
	for managerID, managerFactory := range secretless.InternalConnectionManagers {
		manager.ConnectionManagers[managerID] = managerFactory()
	}
	managerNames := reflect.ValueOf(manager.ConnectionManagers).MapKeys()
	log.Printf("- ConnectionManagers: %v", managerNames)

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
func (manager *Manager) LoadLibraryPlugins(path string, config config.Config) error {
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

		// Load managers
		if err := _LoadConnectionManagers(manager, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load providers
		if err := _LoadProvidersFromPlugin(manager, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load handlers
		if err := _LoadHandlersFromPlugin(manager, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load listeners
		if err := _LoadListenersFromPlugin(manager, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}
	}

	return nil
}

// InitializeConnectionManagers loops through the connection managers and initializes them
func (manager *Manager) InitializeConnectionManagers(c config.Config) error {
	log.Println("Initializing managers...")
	for managerID, connManagerPlugin := range manager.ConnectionManagers {
		err := connManagerPlugin.Initialize(c, manager._ReloadConfig)
		if err != nil {
			log.Printf("Failed to initialize manager '%s': %s\n", managerID, err.Error())

			// TODO: Decide if manager initialization erroring is a fatal error.
			//       For now we treat it as a non-fatal error.
			delete(manager.ConnectionManagers, managerID)
			continue
		}
		log.Printf("Manager '%s' initialized.", managerID)
	}
	log.Println("Initializing managers done.")

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

// Shutdown loops through the connection managers to call Shutdown
func (manager *Manager) Shutdown() {
	for _, connectionManager := range manager.ConnectionManagers {
		connectionManager.Shutdown()
	}
}
