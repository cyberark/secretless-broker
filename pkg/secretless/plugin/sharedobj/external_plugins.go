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

// rawPlugin defines an interface that includes the only method of a Go plugin
// that we care about. This provides an abstraction layer that we can use
// to mock Go plugins.
type rawPlugin interface {
	Lookup(string) (go_plugin.Symbol, error)
}

// ExternalPluginLookupFunc returns all available external plugins.
type ExternalPluginLookupFunc func(
	pluginDir string,
	checksumfile string,
	logger log.Logger,
) (plugin2.AvailablePlugins, error)

// DirectoryPluginLookupFunc returns all available external plugins.
type DirectoryPluginLookupFunc func(
	pluginDir string,
	checksumfile string,
	logger log.Logger,
) (map[string]rawPlugin, error)

// loadPluginsFromDir loads all plugins from a given directory and returns
// a map of TCP and HTTP connector Plugin structs.
func loadPluginsFromDir(
	pluginDir string,
	checksumsFile string,
	logger log.Logger,
) (map[string]rawPlugin, error) {

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
) (map[string]rawPlugin, error) {

	rawPlugins := map[string]rawPlugin{}
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

		logger.Debugf("Adding '%s' as a plugin...", fileName)

		rawPlugins[fileName[:len(fileName)-3]] = pluginObj
	}

	return rawPlugins, nil
}

// ExternalPlugins is used to enumerate all externally-available plugins in a given
// directory.
func ExternalPlugins(
	pluginDir string,
	checksumsFile string,
	logger log.Logger,
) (plugin2.AvailablePlugins, error) {

	return ExternalPluginsWithOptions(
		pluginDir,
		checksumsFile,
		loadPluginsFromDir,
		logger,
	)
}

// checkExternalPluginIDConflicts asserts that a given plugin ID is not
// already being used by an external plugin that has been loaded.
func checkExternalPluginIDConflicts(
	pluginID string,
	pluginType string,
	externalPlugins Plugins,
	logger log.Logger,
) {
	conflictMessage := "%s plugin ID '%s' conflicts with external plugin"

	// Check whether this ID is a duplicate of a loaded external HTTP plugin.
	if _, ok := externalPlugins.HTTPPluginsByID[pluginID]; ok {
		logger.Panicf(conflictMessage, pluginType, pluginID)
	}
	// Check whether this ID is a duplicate of a loaded external TCP plugin.
	if _, ok := externalPlugins.TCPPluginsByID[pluginID]; ok {
		logger.Panicf(conflictMessage, pluginType, pluginID)
	}
}

// ExternalPluginsWithOptions is used to enumerate all externally-available
// plugins in a specified directory to the clients of this method with the
// additional option of being able to specify the lookup function.
func ExternalPluginsWithOptions(
	pluginDir string,
	checksumsFile string,
	getRawPlugins DirectoryPluginLookupFunc,
	logger log.Logger,
) (plugin2.AvailablePlugins, error) {

	rawPlugins, err := getRawPlugins(pluginDir, checksumsFile, logger)
	if err != nil {
		return nil, err
	}

	plugins := NewPlugins()

	for rawPluginName, rawPlugin := range rawPlugins {
		logger.Debugf("Loading plugin '%s'...", rawPluginName)

		logPluginLoadError := func(err error) {
			logger.Errorf("failed to load plugin '%s': %s", rawPluginName, err)
		}

		pluginType, pluginID, err := parsePluginMetadata(rawPlugin, rawPluginName)
		if err != nil {
			logPluginLoadError(err)
			continue
		}

		checkExternalPluginIDConflicts(pluginID, pluginType, plugins, logger)

		switch pluginType {
		case "connector.http":
			httpPluginSym, err := symbolFromName(rawPlugin, "GetHTTPPlugin")
			if err != nil {
				logPluginLoadError(err)
				continue
			}

			httpPluginFunc, ok := httpPluginSym.(func() http.Plugin)
			if !ok {
				logPluginLoadError(errors.New("GetHTTPPlugin couldn't be cast to expected type"))
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
				logPluginLoadError(errors.New("GetTCPPlugin couldn't be cast to expected type"))
				continue
			}

			plugins.TCPPluginsByID[pluginID] = tcpPluginFunc()
		default:
			err = fmt.Errorf("PluginInfo['type'] of '%s' not supported", pluginType)
			logPluginLoadError(err)
			continue
		}
		logger.Infof("Plugin %s/%s loaded", pluginType, pluginID)
	}
	return &plugins, nil
}

func symbolFromName(
	rawPlugin rawPlugin,
	symbolName string,
) (go_plugin.Symbol, error) {

	symbol, err := rawPlugin.Lookup(symbolName)
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
	rawPlugin rawPlugin,
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
