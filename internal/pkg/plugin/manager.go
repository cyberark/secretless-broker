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
	Plugins []*Plugin
}

var _singleton *PluginManager
var _once sync.Once

func GetManager() *PluginManager {
	_once.Do(func() {
		_singleton = &PluginManager{
			Plugins: make([]*Plugin, 0),
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

		loadedPlugin := &Plugin{}
		loadedPlugin._funcInitialize, err = p.Lookup("Initialize")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcCreateListener, err = p.Lookup("CreateListener")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcNewConnection, err = p.Lookup("NewConnection")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcCloseConnection, err = p.Lookup("CloseConnection")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcCreateHandler, err = p.Lookup("CreateHandler")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcDestroyHandler, err = p.Lookup("DestroyHandler")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcResolveVariable, err = p.Lookup("ResolveVariable")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcClientData, err = p.Lookup("ClientData")
		if err != nil {
			log.Println(err)
			continue
		}

		loadedPlugin._funcServerData, err = p.Lookup("ServerData")
		if err != nil {
			log.Println(err)
			continue
		}

		m.Plugins = append(m.Plugins, loadedPlugin)
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
func (m *PluginManager) NewConnection(c net.Conn) {
	for _, plugin := range m.Plugins {
		plugin.NewConnection(c)
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
