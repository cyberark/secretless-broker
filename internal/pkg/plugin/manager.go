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
	"strings"
	"sync"
	"syscall"

	"github.com/conjurinc/secretless/internal/app/secretless"
	secretlessPkg "github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

var _SupportedFileSuffixes = []string{".so"}

func _IsDynamicLibrary(file os.FileInfo) bool {
	hasSupportedSuffix := false
	fileName := file.Name()
	for _, suffix := range _SupportedFileSuffixes {
		if strings.HasSuffix(fileName, suffix) {
			hasSupportedSuffix = true
			break
		}
	}
	return hasSupportedSuffix
}

type PluginManager struct {
	ConnectionManagers map[string]plugin_v1.ConnectionManager
	HandlerFactories   map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler
	ListenerFactories  map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener
	Proxy              secretless.Proxy
}

var _singleton *PluginManager
var _once sync.Once

func GetManager() *PluginManager {
	_once.Do(func() {
		_singleton = &PluginManager{
			ConnectionManagers: make(map[string]plugin_v1.ConnectionManager),
			HandlerFactories:   make(map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler),
			ListenerFactories:  make(map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener),
		}
	})

	return _singleton
}

func (m *PluginManager) RegisterSignalHandlers() {
	log.Println("Registering signal listeners...")
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

		m.Shutdown()

		log.Printf("Exiting...")
		os.Exit(0)
	}()
	log.Println("Signal listeners registered")
}

// _IsSupportedPluginApiVersion returns a boolean indicating if we support
// the interface API version that plugin is advertising
// TODO: Support a list of supported versions
func _IsSupportedPluginApiVersion(version string) bool {
	return version == "0.0.5"
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
func _LoadConnectionManagers(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	rawManagerPluginsFunc, err := pluginObj.Lookup("GetConnectionManagers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	connManagerPluginsFunc, ok := rawManagerPluginsFunc.(func() map[string]func() plugin_v1.ConnectionManager)
	if !ok {
		return errors.New("ERROR! Could not cast GetConnectionManagers to proper type!")
	}
	connManagerPlugins := connManagerPluginsFunc()

	for managerId, connManagerPluginFactory := range connManagerPlugins {
		connManagerPlugin := connManagerPluginFactory()
		pluginManager.ConnectionManagers[managerId] = connManagerPlugin
		log.Printf("Manager '%s' added from plugin %s", managerId, pluginName)
	}

	return nil
}

// _LoadHandlersFromPlugin appends all handlers from the plugin to the pluginManager
func _LoadHandlersFromPlugin(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	rawHandlerPluginsFunc, err := pluginObj.Lookup("GetHandlers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	handlerPluginsFunc, ok := rawHandlerPluginsFunc.(func() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler)
	if !ok {
		return errors.New("ERROR! Could not cast GetcwHandlers to proper type!")
	}
	handlerPlugins := handlerPluginsFunc()
	for handlerId, handlerFactory := range handlerPlugins {
		pluginManager.HandlerFactories[handlerId] = handlerFactory
		log.Printf("Handler factory '%s' added from plugin %s", handlerId, pluginName)
	}

	return nil
}

// _LoadListenersFromPlugin appends all listeners from the plugin to the pluginManager
func _LoadListenersFromPlugin(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	rawListenerPluginsFunc, err := pluginObj.Lookup("GetListeners")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	listenerPluginsFunc, ok := rawListenerPluginsFunc.(func() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener)
	if !ok {
		return errors.New("ERROR! Could not cast GetListeners to proper type!")
	}
	listenerPlugins := listenerPluginsFunc()
	for listenerId, listenerFactory := range listenerPlugins {
		pluginManager.ListenerFactories[listenerId] = listenerFactory
		log.Printf("Listener factory '%s' added from plugin %s", listenerId, pluginName)
	}

	return nil
}

func (pluginManager *PluginManager) _RunHandler(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler {
	// Ensure that we have this handler
	if _, ok := pluginManager.HandlerFactories[id]; !ok {
		log.Panicf("Error! Unrecognized handler id '%s'", id)
	}

	return pluginManager.HandlerFactories[id](options)
}

func (pluginManager *PluginManager) _RunListener(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener {
	// Ensure that we have this listener
	if _, ok := pluginManager.ListenerFactories[id]; !ok {
		log.Panicf("Error! Unrecognized listener id '%s'")
	}

	return pluginManager.ListenerFactories[id](options)
}

// Run is the main wait loop once we load all the plugins
func (pluginManager *PluginManager) Run(configuration config.Config) error {
	pluginManager.InitializeConnectionManagers(configuration)

	pluginManager.Proxy = secretless.Proxy{
		Config:          configuration,
		EventNotifier:   pluginManager,
		RunHandlerFunc:  pluginManager._RunHandler,
		RunListenerFunc: pluginManager._RunListener,
	}

	pluginManager.Proxy.Run()

	return nil
}

// LoadInternalPlugins loads all handlers/listeners/managers that are included in secretless by default
func (pluginManager *PluginManager) LoadInternalPlugins(config config.Config) error {
	log.Println("Enumerating internal plugins...")

	for managerId, managerFactory := range secretless.InternalConnectionManagers {
		pluginManager.ConnectionManagers[managerId] = managerFactory()
		log.Printf("Manager factory '%s' added.", managerId)
	}

	for handlerId, handlerFactory := range secretless.InternalHandlers {
		pluginManager.HandlerFactories[handlerId] = handlerFactory
		log.Printf("Handler factory '%s' added.", handlerId)
	}

	for listenerId, listenerFactory := range secretless.InternalListeners {
		pluginManager.ListenerFactories[listenerId] = listenerFactory
		log.Printf("Listener factory '%s' added.", listenerId)
	}

	log.Println("Completed loading internal plugins.")
	return nil
}

// LoadLibraryPlugins loads all shared object files in `path`
func (m *PluginManager) LoadLibraryPlugins(path string, config config.Config) error {
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
		rawPluginApiVersion, err := pluginObj.Lookup("PluginApiVersion")
		if err != nil {
			log.Println(err)
			continue
		}
		pluginApiVersion := *rawPluginApiVersion.(*string)

		// Bail if plugin interface API is an unsupported version
		log.Printf("Trying to load %s (API v%s)...", file.Name(), pluginApiVersion)
		if !_IsSupportedPluginApiVersion(pluginApiVersion) {
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
		if err := _LoadConnectionManagers(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load handlers
		if err := _LoadHandlersFromPlugin(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load listeners
		if err := _LoadListenersFromPlugin(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}
	}

	return nil
}

func (m *PluginManager) InitializeConnectionManagers(c config.Config) error {
	log.Println("Initializing managers...")
	for managerId, connManagerPlugin := range m.ConnectionManagers {
		err := connManagerPlugin.Initialize(c)
		if err != nil {
			log.Printf("Failed to initialize manager '%s': %s\n", managerId, err.Error())

			// TODO: Decide if manager initialization erroring is a fatal error.
			//       For now we treat it as a non-fatal error.
			delete(m.ConnectionManagers, managerId)
			continue
		}
		log.Printf("Manager '%s' initialized.", managerId)
	}
	log.Println("Initializing managers done.")

	return nil
}

func (m *PluginManager) CreateListener(l plugin_v1.Listener) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.CreateListener(l)
	}
}
func (m *PluginManager) NewConnection(l plugin_v1.Listener, c net.Conn) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.NewConnection(l, c)
	}
}
func (m *PluginManager) CloseConnection(c net.Conn) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.CloseConnection(c)
	}
}
func (m *PluginManager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.CreateHandler(h, c)
	}
}
func (m *PluginManager) DestroyHandler(h plugin_v1.Handler) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.DestroyHandler(h)
	}
}
func (m *PluginManager) ResolveVariable(p secretlessPkg.Provider, id string, value []byte) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.ResolveVariable(p, id, value)
	}
}
func (m *PluginManager) ClientData(c net.Conn, buf []byte) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.ClientData(c, buf)
	}
}
func (m *PluginManager) ServerData(c net.Conn, buf []byte) {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.ServerData(c, buf)
	}
}
func (m *PluginManager) Shutdown() {
	for _, managerPlugin := range m.ConnectionManagers {
		managerPlugin.Shutdown()
	}
}
