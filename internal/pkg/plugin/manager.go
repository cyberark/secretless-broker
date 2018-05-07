package plugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"plugin"
	"strings"
	"sync"

	"github.com/conjurinc/secretless/pkg/secretless"
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

// LoadPlugins loads all shared object files in `path`
func (m *PluginManager) LoadPlugins(path string) error {
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
			m.Plugins = append(m.Plugins, loadedPlugin.(pluginPkg.Plugin))
		} else {
			log.Printf("Failed to load plugin %s\n", file.Name())
			continue
		}
	}

	m.Initialize()

	return nil
}

func (m *PluginManager) Initialize() {
	for _, plugin := range m.Plugins {
		plugin.Initialize()
	}
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
func (m *PluginManager) ClientData(buf []byte) {
	for _, plugin := range m.Plugins {
		plugin.ClientData(buf)
	}
}
func (m *PluginManager) ServerData(buf []byte) {
	for _, plugin := range m.Plugins {
		plugin.ServerData(buf)
	}
}
