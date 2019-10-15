package entrypoint

import (
	"fmt"
	"log"
	"os"

	"github.com/cyberark/secretless-broker/internal/configurationmanagers/configfile"
	secretlessLog "github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/eventnotifier"
	"github.com/cyberark/secretless-broker/internal/profile"
	"github.com/cyberark/secretless-broker/internal/proxyservice"
	"github.com/cyberark/secretless-broker/internal/signal"
	"github.com/cyberark/secretless-broker/internal/util"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj"
)

// SecretlessOptions holds the command line flag information that Service was
// started with.
type SecretlessOptions struct {
	ConfigFile          string
	ConfigManagerSpec   string
	DebugEnabled        bool
	FsWatchEnabled      bool
	PluginChecksumsFile string
	PluginDir           string
	ProfilingMode       string
	ShowVersion         bool
}

// StartSecretless method is the main entry point into the broker after the CLI
// flags have been parsed
func StartSecretless(params *SecretlessOptions) {
	showVersion(params.ShowVersion)

	// Construct the deps of Service

	// Coordinates processes interested in exit signals
	exitListener := signal.NewExitListener()

	logger := secretlessLog.New(params.DebugEnabled)
	evtNotifier := eventnotifier.New(nil)
	availPlugins, err := sharedobj.AllAvailablePlugins(
		params.PluginDir,
		params.PluginChecksumsFile,
		logger,
	)

	if err != nil {
		log.Fatalln(err)
	}

	// Optional Performance Profiling
	handlePerformanceProfiling(params.ProfilingMode, exitListener)

	configChangedChan, err := newConfigChangeChan(
		params.ConfigFile,
		params.ConfigManagerSpec,
		params.FsWatchEnabled,
	)
	if err != nil {
		log.Fatalln(err)
	}

	// Health check: Initialized
	util.SetAppInitializedFlag()
	util.SetAppIsLive(false)

	configChangedFunc := func(cfg v2.Config) {
		util.SetAppIsLive(false)
		// Start Services
		allServices := proxyservice.NewProxyServices(cfg, availPlugins, logger, evtNotifier)
		exitListener.AddHandler(func() {
			fmt.Println("Received a stop signal")
			err := allServices.Stop()
			if err != nil {
				// Log but but allow cleanup of other subscribers to continue.
				log.Println(err)
			}

			os.Exit(0)
		})

		err = allServices.Start()
		if err != nil {
			log.Fatalln(err)
		}

		// Health check: Live
		util.SetAppIsLive(true)
		exitListener.Wait()
	}

	logger.Info("Waiting for configuration...")
	cfg := <-configChangedChan

	logger.Debug("Got new configuration")
	configChangedFunc(cfg)

	logger.Info("Exiting...")
}

func newConfigChangeChan(
	cfgFile string,
	cfgManagerSpec string,
	fsWatchEnabled bool,
) (<-chan v2.Config, error) {

	if cfgManagerSpec != "configfile" {
		return nil, fmt.Errorf("only 'configfile' configuration manager is supported")
	}

	return configfile.NewManager(cfgFile, fsWatchEnabled)
}

func showVersion(showAndExit bool) {
	if showAndExit {
		fmt.Printf("secretless-broker v%s\n", secretless.FullVersionName)
		os.Exit(0)
	}
	log.Printf("Secretless v%s starting up...", secretless.FullVersionName)
}

// handlePerformanceProfiling starts a performance profiling, and sets up an
// os.Signal listener that will automatically call Stop() on the profile
// when an system halt is raised.
func handlePerformanceProfiling(profileType string, exitSignals signal.ExitListener) {
	// No profiling was requested
	if profileType == "" {
		return
	}

	// Validate requested type
	if err := profile.ValidateType(profileType); err != nil {
		log.Fatalln(err)
	}

	// Start profiling
	perfProfile := profile.New(profileType)
	exitSignals.AddHandler(func() {
		_ = perfProfile.Stop()
	})

	err := perfProfile.Start()
	if err != nil {
		log.Fatalln(err)
	}
}
