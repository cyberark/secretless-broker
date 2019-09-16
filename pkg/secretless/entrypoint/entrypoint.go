package entrypoint

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	secretlessLog "github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/event_notifier"
	"github.com/cyberark/secretless-broker/internal/profile"
	"github.com/cyberark/secretless-broker/internal/proxy_service"
	"github.com/cyberark/secretless-broker/internal/signal"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// SecretlessOptions holds the command line flag information that StartProxyServices was started
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

	verifyPlugins(params.PluginDir, params.PluginChecksumsFile)

	// Construct the deps of StartProxyServices
	cfg := readConfig(params.ConfigFile)
	logger := secretlessLog.New(params.DebugEnabled)
	evtNotifier := event_notifier.New(nil)
	availPlugins := &proxy_service.AvailPluginStub{}

	// Prepare StartProxyServices
	secretless := proxy_service.NewStartProxyServices(cfg, availPlugins, logger, evtNotifier)
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

// verifyPlugins is responsible only for the verification of the plugin
// checksum file, and warnings when no file is present.  Even though it
// currently delegates to VerifyPluginChecksums, it is not concerned with
// the validated files that function returns.
func verifyPlugins(pluginDir string, checksumFile string) {
	// No external plugin loading was requested
	if pluginDir == "" {
		return
	}

	// Read the requested files
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		log.Fatalln(err)
	}

	// No files to verify, just return
	if len(files) == 0 {
		return
	}

	// Warn if we're loading plugins without a checksum
	if checksumFile == "" {
		log.Println("WARN: No PluginChecksumsFile provided - plugin tampering" +
			" is possible!")
		return
	}

	if _, err = plugin.VerifyPluginChecksums(pluginDir, checksumFile); err != nil {
		log.Fatalln(err)
	}
}
