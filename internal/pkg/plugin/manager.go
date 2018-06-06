package plugin

import (
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
	pluginPkg "github.com/conjurinc/secretless/pkg/secretless/plugin"
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
	Plugins []pluginPkg.Plugin
}

var _singleton *PluginManager
var _once sync.Once

func GetManager() *PluginManager {
	_once.Do(func() {
		_singleton = &PluginManager{
			Plugins: make([]pluginPkg.Plugin, 0),
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

		p, err := plugin.Open(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		getPlugin, err := p.Lookup("GetPlugin")
		if err != nil {
			log.Println(err)
			continue
		}

		if getPluginFunc, ok := getPlugin.(func() pluginPkg.Plugin); ok {
			loadedPlugin := getPluginFunc()

			err = loadedPlugin.Initialize(config)
			if err != nil {
				log.Printf("%s: %s\n", file.Name(), err.Error())
				continue
			}

			m.Plugins = append(m.Plugins, loadedPlugin.(pluginPkg.Plugin))
		} else {
			log.Printf("Failed to load plugin %s\n", file.Name())
			continue
		}
	}

	return nil
}

func (m *PluginManager) Initialize(c config.Config) error {
	for _, plugin := range m.Plugins {
		plugin.Initialize(c)
	}
	return nil
}

func (m *PluginManager) CreateListener(l secretless.Listener) {
	for _, plugin := range m.Plugins {
		plugin.CreateListener(l)
	}
}
func (m *PluginManager) NewConnection(l secretless.Listener, c net.Conn) {
	for _, plugin := range m.Plugins {
		plugin.NewConnection(l, c)
	}
}
func (m *PluginManager) CloseConnection(c net.Conn) {
	for _, plugin := range m.Plugins {
		plugin.CloseConnection(c)
	}
}
func (m *PluginManager) CreateHandler(l secretless.Listener, h secretless.Handler) {
	for _, plugin := range m.Plugins {
		plugin.CreateHandler(l, h)
	}
}
func (m *PluginManager) DestroyHandler(h secretless.Handler) {
	for _, plugin := range m.Plugins {
		plugin.DestroyHandler(h)
	}
}
func (m *PluginManager) ResolveVariable(p secretless.Provider, id string, value []byte) {
	for _, plugin := range m.Plugins {
		plugin.ResolveVariable(p, id, value)
	}
}
func (m *PluginManager) ClientData(c net.Conn, buf []byte) {
	for _, plugin := range m.Plugins {
		plugin.ClientData(c, buf)
	}
}
func (m *PluginManager) ServerData(c net.Conn, buf []byte) {
	for _, plugin := range m.Plugins {
		plugin.ServerData(c, buf)
	}
}
func (m *PluginManager) Shutdown() {
	for _, plugin := range m.Plugins {
		plugin.Shutdown()
	}
}
