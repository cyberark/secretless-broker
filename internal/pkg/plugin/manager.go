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

	"github.com/conjurinc/secretless/pkg/secretless"
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
	Listeners []plugin_v1.Listener
	Handlers  []plugin_v1.Handler
	Managers  []plugin_v1.Manager
}

var _singleton *PluginManager
var _once sync.Once

func GetManager() *PluginManager {
	_once.Do(func() {
		_singleton = &PluginManager{
			Listeners: make([]plugin_v1.Listener, 0),
			Handlers:  make([]plugin_v1.Handler, 0),
			Managers:  make([]plugin_v1.Manager, 0),
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
	return version == "0.0.2"
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

// _LoadManagers appends all managers from the plugin to the pluginManager
func _LoadManagers(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	rawManagerPluginsFunc, err := pluginObj.Lookup("GetManagers")
	if err != nil {
		return err
	}

	// TODO: Handle different interface versions
	managerPluginsFunc, ok := rawManagerPluginsFunc.(func() map[string]plugin_v1.Manager)
	if !ok {
		return errors.New("ERROR! Could not cast GetManagers to proper type!")
	}
	managerPlugins := managerPluginsFunc()

	for managerPluginName, managerPlugin := range managerPlugins {
		log.Printf("%s: Appending manager '%s'...", pluginName, managerPluginName)
		err = managerPlugin.Initialize(config)
		if err != nil {
			log.Printf("%s: Failed to load manager '%s': %s\n", pluginName,
				managerPluginName,
				err.Error())
			return err
		}
		log.Printf("%s: Appending manager '%s' OK!", pluginName, managerPluginName)

		pluginManager.Managers = append(pluginManager.Managers,
			managerPlugin.(plugin_v1.Manager))
	}

	return nil
}

// _LoadHandlers appends all handlers from the plugin to the pluginManager
func _LoadHandlers(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	log.Println("WARN: Loading of handlers from a plugin is a NOOP at this time!")

	return nil
}

// _LoadListeners appends all listeners from the plugin to the pluginManager
func _LoadListeners(pluginManager *PluginManager, config config.Config,
	pluginObj *plugin.Plugin, pluginName string) error {

	log.Println("WARN: Loading of listeners from a plugin is a NOOP at this time!")

	return nil
}

// LoadPlugins loads all shared object files in `path`
func (m *PluginManager) LoadPlugins(path string, config config.Config) error {
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
		if err := _LoadManagers(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load handlers
		// TODO: Actually load handlers
		if err := _LoadHandlers(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}

		// Load listeners
		// TODO: Actually load listeners
		if err := _LoadListeners(m, config, pluginObj, file.Name()); err != nil {
			// Log the error but try to load other plugins
			log.Println(err)
		}
	}

	return nil
}

func (m *PluginManager) Initialize(c config.Config) error {
	// TODO: Figure out if handlers/listeners need to initialize too
	for _, managerPlugin := range m.Managers {
		managerPlugin.Initialize(c)
	}

	return nil
}

func (m *PluginManager) CreateListener(l plugin_v1.Listener) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.CreateListener(l)
	}
}
func (m *PluginManager) NewConnection(l plugin_v1.Listener, c net.Conn) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.NewConnection(l, c)
	}
}
func (m *PluginManager) CloseConnection(c net.Conn) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.CloseConnection(c)
	}
}
func (m *PluginManager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.CreateHandler(h, c)
	}
}
func (m *PluginManager) DestroyHandler(h plugin_v1.Handler) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.DestroyHandler(h)
	}
}
func (m *PluginManager) ResolveVariable(p secretless.Provider, id string, value []byte) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.ResolveVariable(p, id, value)
	}
}
func (m *PluginManager) ClientData(c net.Conn, buf []byte) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.ClientData(c, buf)
	}
}
func (m *PluginManager) ServerData(c net.Conn, buf []byte) {
	for _, managerPlugin := range m.Managers {
		managerPlugin.ServerData(c, buf)
	}
}
func (m *PluginManager) Shutdown() {
	for _, managerPlugin := range m.Managers {
		managerPlugin.Shutdown()
	}
}
