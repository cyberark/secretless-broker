package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cyberark/summon/secretsyml"
	"github.com/conjurinc/secretless/internal/pkg/provider"
)

// Subcommand defines the input needed to run Summon.
type Subcommand struct {
	Args        []string
	Providers   []provider.Provider
	SecretsMap  secretsyml.SecretsMap
	TempFactory *TempFactory
}

func findProvider(providers []provider.Provider, secretSpec secretsyml.SecretSpec) (provider.Provider, error) {
	if len(providers) == 1 {
		return providers[0], nil
	}
	return nil, fmt.Errorf("findProviders is not implemented for multiple providers")
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
func resolveVariables(providers []provider.Provider, secretsMap secretsyml.SecretsMap) (result map[string]string, err error) {
	result = make(map[string]string)
	for key, spec := range secretsMap {
		var value string
		if spec.IsVar() {
			var provider provider.Provider
			if provider, err = findProvider(providers, spec); err != nil {
				return
			}
			var valueBytes []byte
			if valueBytes, err = provider.Value(spec.Path); err != nil {
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
// of an environment populated with secret values.
//
// It returns the command stdout, and sderr if any. The command stdout and stderr
// are also written to this process' stdout and stderr.
func runSubcommand(command []string, env []string) (stdout string, err error) {
	var stdOut bytes.Buffer

	runner := exec.Command(command[0], command[1:]...)
	runner.Stdin = os.Stdin
	runner.Stdout = io.MultiWriter(os.Stdout, &stdOut)
	runner.Stderr = os.Stderr
	runner.Env = env

	err = runner.Run()
	stdout = stdOut.String()

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
func (sc *Subcommand) Run() (stdout string, err error) {
	var env []string
	var secrets map[string]string

	if sc.TempFactory == nil {
		tempFactory := NewTempFactory("")
		sc.TempFactory = &tempFactory
	}
	defer sc.TempFactory.Cleanup()

	if secrets, err = resolveVariables(sc.Providers, sc.SecretsMap); err != nil {
		return
	}
	if env, err = buildEnvironment(secrets, sc.SecretsMap, sc.TempFactory); err != nil {
		return
	}

	stdout, err = runSubcommand(sc.Args, append(os.Environ(), env...))
	return
}
