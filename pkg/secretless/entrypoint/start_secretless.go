package entrypoint

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cyberark/secretless-broker/internal"
	secretlessLog "github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/event_notifier"
	"github.com/cyberark/secretless-broker/internal/profile"
	"github.com/cyberark/secretless-broker/internal/signal"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// CLIParams holds the command line flag information that Secretless was started
// with.
type CLIParams struct {
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
func StartSecretless(params *CLIParams) {

	ShowVersion(params.ShowVersion)

	VerifyPlugins(params.PluginDir, params.PluginChecksumsFile)

	// Construct the deps of Secretless
	cfg := ReadConfig(params.ConfigFile)
	logger := secretlessLog.New(params.DebugEnabled)
	evtNotifier := event_notifier.New(nil)
	availPlugins := &internal.AvailPluginStub{}

	// Prepare Secretless
	secretless := internal.NewSecretless(cfg, availPlugins, logger, evtNotifier)
	signal.StopOnExitSignal(secretless)

	HandlePerformanceProfiling(params.ProfilingMode)

	secretless.Start()
}

func ReadConfig(cfgFile string) v2.Config {
	// TODO: Add back in CRD / generalized config option
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}

func ShowVersion(showVersionAndExit bool) {
	if showVersionAndExit {
		fmt.Printf("secretless-broker v%s\n", secretless.FullVersionName)
		os.Exit(0)
	}
	log.Printf("Secretless v%s starting up...", secretless.FullVersionName)
}

// HandlePerformanceProfiling starts a perfomance profiling, and sets up an
// os.Signal listener that will automatically call Stop() on the profile
// when an system halt is raised.
func HandlePerformanceProfiling(profileType string) {
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

// VerifyPlugins is responsible only for the verification of the plugin
// checksum file, and warnings when no file is present.  Even though it
// currently delegates to VerifyPluginChecksums, it is not concerned with
// the validated files that function returns.
func VerifyPlugins(pluginDir string, checksumFile string) {
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
