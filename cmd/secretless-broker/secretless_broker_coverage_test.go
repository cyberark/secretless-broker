package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/cyberark/secretless-broker/pkg/secretless/entrypoint"
	"github.com/stretchr/testify/assert"
)

// TestCoverage runs the Secretless Broker executable wrapped in a
// Go test so that the Go test infrastructure can collect code coverage stats
// for the Secretless Broker e.g. while running Docker-Compose-based
// integration tests. This Go test and the Secretless Broker code can be
// compiled into a special Secretless Broker test binary that includes
// code coverage instrumentation using the following Go test compile command:
//
//        go test -c -coverpkg="./..." ./cmd/secretless-broker
//
// rather than the usual `go build ...` compile command. (See
// 'Dockerfile.coverage' in the root of this GitHub repository to see how this
// compilation is done and built into a Docker image).
//
// Using this specially-instrumented Secretless Broker test binary, the
// Go test with its embedded Secretless Broker functionality would be invoked
// e.g. using the following command (note that standard Go test parameters
// must be prefixed with '-test.' for images compiled via 'go test -c ...'):
//
//        export SB_RUN_COVERAGE=true
//        /usr/local/bin/secretless-broker \
//            -test.v \
//            -test.run ^TestCoverage$$ \
//            -test.coverprofile=/test-coverage/cover.out"
//
// Normally, the Secretless Broker binary can be passed optional command line
// parameters, for example:
//
//        secretless-broker -f /secretless-test.yml -debug
//
// However, in the case of the Secretless Broker test binary, command line
// parameters are processed by the Go test infrastructure, and therefore any
// of the usual Secretless Broker command line parameters will be rejected
// and cause a test failure.
//
// To get around this limitation, the TestCoverage test will
// look for equivalent environment variable settings for the usual Secretless
// Broker command line parameters, and pass those settings to the Secretless
// Broker code. These environment variable settings include:
//
//       Environment Variable      Equivalent Command Line Parameter
//       --------------------      ---------------------------------
//       SB_CONFIG_FILE            -f <config-file>
//       SB_PROFILING_MODE         -profile [cpu|memory]
//       SB_DEBUG_ENABLED          -debug
//       SB_CONFIG_MANAGER         -config-mgr <config-mgr-spec>
//       SB_FS_WATCH_ENABLED       -watch
//       SB_PLUGIN_DIR             -p <plugin-directory>
//       SB_PLUGIN_CHECKSUM_FILE   -s <plugin-checksum-file>

func TestCoverage(t *testing.T) {
	skipForUT(t)
	params, err := getSecretlessOptions()
	assert.NoError(t, err)
	entrypoint.StartSecretless(params)
}

// skipForUT skips the TestCoverage test unless the SB_RUN_COVERAGE
// environment variable is set, which should only be set for integration
// testing.
func skipForUT(t *testing.T) {
	if os.Getenv("SB_RUN_COVERAGE") == "" {
		t.Skip("Skipping TestCoverage(). This is intended for integration tests, not unit tests")
	}
}

func getSecretlessOptions() (*entrypoint.SecretlessOptions, error) {
	// Use CmdLineParams() to get the default Secretless Broker settings.
	// There shouldn't be any Secretless Broker command line settings that
	// are set explicitly, since any such command line settings would not
	// be understood by Go test, causing the Go test run to exit with error.
	params := CmdLineParams()

	// If any corresponding Secretless Broker environment variables are set,
	// override the default command line parameters.
	mappings := []struct {
		envVarName string
		param      interface{}
	}{
		{"SB_CONFIG_FILE", &params.ConfigFile},
		{"SB_PROFILING_MODE", &params.ProfilingMode},
		{"SB_DEBUG_ENABLED", &params.DebugEnabled},
		{"SB_CONFIG_MANAGER", &params.ConfigManagerSpec},
		{"SB_FS_WATCH_ENABLED", &params.FsWatchEnabled},
		{"SB_PLUGIN_DIR", &params.PluginDir},
		{"SB_PLUGIN_CHECKSUM_FILE", &params.PluginChecksumsFile},
	}
	for _, m := range mappings {
		if err := overrideOption(m.envVarName, m.param); err != nil {
			return nil, err
		}
	}
	// Avoid calls to os.Exit() so that Go test can report coverage
	params.GracefulExitEnabled = true
	return params, nil
}

func overrideOption(envVarName string, dest interface{}) error {
	value, exists := os.LookupEnv(envVarName)
	if exists {
		fmt.Printf("Overriding config with env var %s=\"%s\"\n", envVarName, value)
		switch t := dest.(type) {
		case *string:
			*dest.(*string) = value
		case *int:
			setting, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			*dest.(*int) = setting
		case *bool:
			setting, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			*dest.(*bool) = setting
		default:
			return fmt.Errorf("unexpected type in env var override: %T", t)
		}
	}
	return nil
}
