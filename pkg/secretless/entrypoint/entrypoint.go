package entrypoint

import (
	"fmt"
	"log"
	"os"

	secretlessLog "github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/eventnotifier"
	"github.com/cyberark/secretless-broker/internal/profile"
	"github.com/cyberark/secretless-broker/internal/proxyservice"
	"github.com/cyberark/secretless-broker/internal/signal"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// SecretlessOptions holds the command line flag information that Service was started
// with.
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
	cfg := readConfig(params.ConfigFile)
	logger := secretlessLog.New(params.DebugEnabled)
	evtNotifier := eventnotifier.New(nil)
	availPlugins, err := plugin.AllAvailablePlugins(
		params.PluginDir,
		params.PluginChecksumsFile,
		logger,
	)

	if err != nil {
		log.Fatalln(err)
	}

	// Prepare Service
	secretless := proxyservice.NewProxyServices(cfg, availPlugins, logger, evtNotifier)
	signal.StopOnExitSignal(secretless)

	handlePerformanceProfiling(params.ProfilingMode)

	secretless.Start()
}

func readConfig(cfgFile string) v2.Config {
	// TODO: Add back in CRD / generalized config option
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}

func showVersion(showAndExit bool) {
	if showAndExit {
		fmt.Printf("secretless-broker v%s\n", secretless.FullVersionName)
		os.Exit(0)
	}
	log.Printf("Secretless v%s starting up...", secretless.FullVersionName)
}

// handlePerformanceProfiling starts a perfomance profiling, and sets up an
// os.Signal listener that will automatically call Stop() on the profile
// when an system halt is raised.
func handlePerformanceProfiling(profileType string) {
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
	signal.StopOnExitSignal(perfProfile)
	perfProfile.Start()
}
