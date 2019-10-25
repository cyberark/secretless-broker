package sharedobj

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	go_plugin "plugin"
	"strings"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	plugin2 "github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// ExternalPluginLookupFunc returns all available external plugins.
type ExternalPluginLookupFunc func(
	pluginDir string,
	checksumfile string,
	logger log.Logger,
) (map[string]*go_plugin.Plugin, error)

// LoadPluginsFromDir loads all plugins from a given directory and returns
// a map of TCP and HTTP connector Plugin structs.
func LoadPluginsFromDir(
	pluginDir string,
	checksumsFile string,
	logger log.Logger,
) (map[string]*go_plugin.Plugin, error) {

	// Missing external plugin folder is a warning not a fatal error
	_, err := os.Stat(pluginDir)
	if os.IsNotExist(err) {
		logger.Warnf(
			"Plugin directory '%s' not found. Ignoring external plugins...",
			pluginDir,
		)
		return nil, nil
	}

	filePaths, err := checkedPlugins(pluginDir, checksumsFile, logger)
	if err != nil {
		return nil, err
	}

	return loadPluginFiles(filePaths, logger)
}

func checkedPlugins(
	pluginDir string,
	checksumsFile string,
	logger log.Logger,
) ([]string, error) {

	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	if checksumsFile == "" {
		logger.Warnln("Plugin hashes were not provided - tampering will not be detectable!")
		return filePaths(pluginDir, files), nil
	}

	// We override file listing if we did a verification to prevent additions
	// to plugins between verification and loading the plugins.
	if files, err = VerifyPluginChecksums(pluginDir, checksumsFile); err != nil {
		logger.Errorln(err)
		return nil, err
	}

	return filePaths(pluginDir, files), nil
}

func filePaths(pluginDir string, files []os.FileInfo) []string {
	filePaths := []string{}
	for _, file := range files {
		filePaths = append(filePaths, path.Join(pluginDir, file.Name()))
	}

	return filePaths
}

func loadPluginFiles(
	filePaths []string,
	logger log.Logger,
) (map[string]*go_plugin.Plugin, error) {

	goPlugins := map[string]*go_plugin.Plugin{}
	for _, filePath := range filePaths {
		fileName := path.Base(filePath)
		if !strings.HasSuffix(fileName, ".so") {
			logger.Warnf("File '%s' ignored as a plugin - missing appropriate extension",
				fileName)
			continue
		}

		// Load shared library object
		pluginObj, err := go_plugin.Open(filePath)
		if err != nil {
			logger.Errorln(err)
			continue
		}

		logger.Infof("Adding '%s' as a plugin...", fileName)

		goPlugins[fileName[:len(fileName)-3]] = pluginObj
	}

	return goPlugins, nil
}

// ExternalPlugins is used to enumerate all externally-available plugins in a given
// directory.
func ExternalPlugins(
	pluginDir string,
	getRawPlugins ExternalPluginLookupFunc,
	logger log.Logger,
	checksumsFile string,
) (plugin2.AvailablePlugins, error) {

	//TODO: Test this

	rawPlugins, err := getRawPlugins(pluginDir, checksumsFile, logger)
	if err != nil {
		return nil, err
	}

	plugins := Plugins{
		HTTPPluginsByID: map[string]http.Plugin{},
		TCPPluginsByID:  map[string]tcp.Plugin{},
	}

	for rawPluginName, rawPlugin := range rawPlugins {
		logger.Infof("Loading plugin '%s'...", rawPluginName)

		logPluginLoadError := func(err error) {
			logger.Errorf(
				"failed to load plugin '%s': %s",
				rawPluginName,
				err,
			)
		}

		pluginType, pluginID, err := parsePluginMetadata(rawPlugin, rawPluginName)
		if err != nil {
			logPluginLoadError(err)
			continue
		}

		switch pluginType {
		case "connector.http":
			httpPluginSym, err := symbolFromName(rawPlugin, "GetHTTPPlugin")
			if err != nil {
				logPluginLoadError(err)
				continue
			}

			httpPluginFunc, ok := httpPluginSym.(func() http.Plugin)
			if !ok {
				err := errors.New("GetHTTPPlugin could not be cast to the expected type")
				logPluginLoadError(err)
				continue
			}

			plugins.HTTPPluginsByID[pluginID] = httpPluginFunc()
		case "connector.tcp":
			tcpPluginSym, err := symbolFromName(rawPlugin, "GetTCPPlugin")
			if err != nil {
				logPluginLoadError(err)
				continue
			}

			tcpPluginFunc, ok := tcpPluginSym.(func() tcp.Plugin)
			if !ok {
				err = errors.New(
					"GetTCPPlugin could not be cast to the expected type",
				)
				logPluginLoadError(err)
				continue
			}

			plugins.TCPPluginsByID[pluginID] = tcpPluginFunc()
		default:
			err = fmt.Errorf(
				"PluginInfo['type'] of '%s' is not supported",
				pluginType,
			)
			logPluginLoadError(err)
			continue
		}

		logger.Warnf("Plugin %s/%s loaded", pluginType, pluginID)
	}

	return &plugins, nil
}

func symbolFromName(
	rawPlugin *go_plugin.Plugin,
	symbolName string,
) (go_plugin.Symbol, error) {

	symbol, err := (*rawPlugin).Lookup(symbolName)
	if err != nil {
		return nil, err
	}

	return symbol, nil
}

func infoField(info map[string]string, fieldName string) (string, error) {
	fieldValue, ok := info[fieldName]
	if !ok {
		err := fmt.Errorf("PluginInfo does not contain '%s' field", fieldName)
		return "", err
	}

	if fieldValue == "" {
		err := fmt.Errorf("PluginInfo['%s'] is blank", fieldName)
		return "", err
	}

	return fieldValue, nil
}

func parsePluginMetadata(
	rawPlugin *go_plugin.Plugin,
	rawPluginName string,
) (pluginType string, pluginID string, err error) {

	pluginInfoSym, err := symbolFromName(rawPlugin, "PluginInfo")
	if err != nil {
		return pluginType, pluginID, err
	}

	pluginInfoFunc, ok := pluginInfoSym.(func() map[string]string)
	if !ok {
		err = errors.New("could not cast PluginInfo to proper type")
		return pluginType, pluginID, err
	}
	pluginInfo := pluginInfoFunc()

	pluginAPIVersion, err := infoField(pluginInfo, "pluginAPIVersion")
	if err != nil {
		return pluginType, pluginID, err
	}

	if pluginAPIVersion != CompatiblePluginAPIVersion {
		err = fmt.Errorf("plugin '%s' (API v%s) is not a supported API version (v%s)",
			rawPluginName, pluginAPIVersion, CompatiblePluginAPIVersion)
		return pluginType, pluginID, err
	}

	pluginType, err = infoField(pluginInfo, "type")
	if err != nil {
		return pluginType, pluginID, err
	}

	pluginID, err = infoField(pluginInfo, "id")
	if err != nil {
		return pluginType, pluginID, err
	}

	// TODO: Verify PluginID charset

	return pluginType, pluginID, nil
}
