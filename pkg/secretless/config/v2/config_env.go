package v2

import (
	"fmt"
	"os"
	"strings"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/go-ozzo/ozzo-validation"
)

// DeleteFileFunc is a function that takes a filename, attempts to delete the
// file, and returns an error if it can't.
type DeleteFileFunc func(name string) error

// FileInfoFunc is a function that takes a filename and returns information
// about that file, or an error if it cannot be found or read.
type FileInfoFunc func(name string) (os.FileInfo, error)

// ConfigEnv represents the runtime environment that will fulfill the services
// requested by the Config.  It has a single public method, Prepare, that
// ensures the runtime environment supports the requested services.
type ConfigEnv struct {
	availPlugins plugin.AvailablePlugins
	deleteFile   DeleteFileFunc
	getFileInfo  FileInfoFunc
	logger       log.Logger
}

// NewConfigEnv creates a new instance of ConfigEnv.
func NewConfigEnv(logger log.Logger, availPlugins plugin.AvailablePlugins) ConfigEnv {
	return NewConfigEnvWithOptions(logger, availPlugins, os.Stat, os.Remove)
}

// NewConfigEnvWithOptions allows injecting all dependencies.  Used for unit
// testing.
func NewConfigEnvWithOptions(
	logger log.Logger,
	availPlugins plugin.AvailablePlugins,
	getFileInfo func(name string) (os.FileInfo, error),
	deleteFile func(name string) error,
) ConfigEnv {
	return ConfigEnv{
		logger:       logger,
		availPlugins: availPlugins,
		getFileInfo:  getFileInfo,
		deleteFile:   deleteFile,
	}
}

// Prepare ensures the runtime environment is prepared to handle the Config's
// service requests. It checks both that the requested connectors exist, and
// that the requested sockets are available, or can be deleted.  If any of these
// checks fail, it will error.
func (c *ConfigEnv) Prepare(cfg Config) error {
	err := c.validateRequestedPlugins(cfg)
	if err != nil {
		return err
	}

	return c.ensureAllSocketsAreDeleted(cfg)
}

// validateRequestedPlugins ensures that the AvailablePlugins can fulfill the
// services requested by the given Config, and return an error if not.
func (c *ConfigEnv) validateRequestedPlugins(cfg Config) error {
	pluginIDs := plugin.AvailableConnectorIDs(c.availPlugins)

	c.logger.Infof(
		"Validating config against available plugins: %s",
		strings.Join(pluginIDs, ","),
	)

	// Convert available plugin IDs to a map, so that we can check if they exist
	// in the loop below using a map lookup rather than a nested loop.
	pluginIDsMap := map[string]bool{}
	for _, p := range pluginIDs {
		pluginIDsMap[p] = true
	}

	errors := validation.Errors{}
	for _, service := range cfg.Services {
		// A plugin ID and a connector name are equivalent.
		pluginExists := pluginIDsMap[service.Connector]
		if !pluginExists {
			errors[service.Name] = fmt.Errorf(
				`missing service connector "%s"`,
				service.Connector,
			)
			continue
		}
	}

	err := errors.Filter()
	if err != nil {
		err = fmt.Errorf("services validation failed: %s", err.Error())
	}

	return err
}

func (c *ConfigEnv) ensureAllSocketsAreDeleted(cfg Config) error {
	errors := validation.Errors{}

	for _, service := range cfg.Services {
		err := c.ensureSocketIsDeleted(service.ListenOn)
		if err != nil {
			errors[service.Name] = fmt.Errorf(
				"socket can't be deleted: %s", service.ListenOn,
			)
		}
	}
	return errors.Filter()
}

func (c *ConfigEnv) ensureSocketIsDeleted(address NetworkAddress) error {
	// If we're not a unix socket address, we don't need to worry about
	// pre-emptive cleanup
	if address.Network() != "unix" {
		return nil
	}

	socketFile := address.Address()
	c.logger.Debugf("Ensuring that the socketfile '%s' is not present...", socketFile)

	// If file is not present, then we are ok to continue.
	// NOTE: os.IsNotExist is a pure function, so does not need to be injected.
	if _, err := c.getFileInfo(socketFile); os.IsNotExist(err) {
		c.logger.Debugf("Socket file '%s' not present. Skipping deletion.", socketFile)
		return nil
	}

	// Otherwise delete the file first
	c.logger.Warnf("Socket file '%s' already present. Deleting...", socketFile)
	err := c.deleteFile(socketFile)
	if err != nil {
		return fmt.Errorf("unable to delete stale socket file '%s'", socketFile)
	}

	return nil
}
