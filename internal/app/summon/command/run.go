package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cyberark/summon/secretsyml"

	plugin_v1 "github.com/conjurinc/secretless-broker/pkg/secretless/plugin/v1"
)

// Subcommand defines the input needed to run Summon.
type Subcommand struct {
	Args        []string
	SecretsMap  secretsyml.SecretsMap
	TempFactory *TempFactory
	Provider    plugin_v1.Provider

	// Set this to an io.Writer to capture stdout from the child process.
	// By default, the child process stdout goes to this process' stdout.
	Stdout io.Writer
}

// buildEnvironment builds the environment strings from the map of secrets values, along with the
// secrets configuration metadata and the temp files location.
func buildEnvironment(secrets map[string]string, secretsMap secretsyml.SecretsMap, tempFactory *TempFactory) (env []string, err error) {
	env = make([]string, len(secrets))
	for key, value := range secrets {
		envvar := formatForEnv(key, value, secretsMap[key], tempFactory)
		env = append(env, envvar)
	}
	return
}

// resolveVariables obtains the value of each requested secret.
func resolveVariables(provider plugin_v1.Provider, secretsMap secretsyml.SecretsMap) (result map[string]string, err error) {
	result = make(map[string]string)
	for key, spec := range secretsMap {
		var value string
		if spec.IsVar() {
			var valueBytes []byte
			if valueBytes, err = provider.GetValue(spec.Path); err != nil {
				return
			}
			value = string(valueBytes)
		} else {
			// If the spec isn't a variable, use its value as-is
			value = spec.Path
		}
		result[key] = value
	}
	return
}

// runSubcommand executes a command with arguments in the context
// of an environment populated with secret values. The command stdout and stderr
// are also written to this process' stdout and stderr.
//
// It returns the command error, if any.
func (sc *Subcommand) runSubcommand(env []string) (err error) {
	command := sc.Args
	runner := exec.Command(command[0], command[1:]...)
	runner.Stdin = os.Stdin
	if sc.Stdout != nil {
		runner.Stdout = sc.Stdout
	} else {
		runner.Stdout = os.Stdout
	}
	runner.Stderr = os.Stderr
	runner.Env = env

	err = runner.Run()

	return
}

// formatForEnv returns a string in %k=%v format, where %k=namespace of the secret and
// %v=the secret value or path to a temporary file containing the secret
func formatForEnv(key string, value string, spec secretsyml.SecretSpec, tempFactory *TempFactory) string {
	if spec.IsFile() {
		fname := tempFactory.Push(value)
		value = fname
	}

	return fmt.Sprintf("%s=%s", key, value)
}

// Run encapsulates the logic of Action without cli Context for easier testing
func (sc *Subcommand) Run() (err error) {
	var env []string
	var secrets map[string]string

	if sc.TempFactory == nil {
		tempFactory := NewTempFactory("")
		sc.TempFactory = &tempFactory
	}
	defer sc.TempFactory.Cleanup()

	if secrets, err = resolveVariables(sc.Provider, sc.SecretsMap); err != nil {
		return
	}
	if env, err = buildEnvironment(secrets, sc.SecretsMap, sc.TempFactory); err != nil {
		return
	}

	return sc.runSubcommand(append(os.Environ(), env...))
}
