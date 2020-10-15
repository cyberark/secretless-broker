package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/cyberark/summon/secretsyml"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
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
func buildEnvironment(
	secrets map[string]string,
	secretsMap secretsyml.SecretsMap,
	tempFactory *TempFactory,
) ([]string, error) {
	env := make([]string, 0, len(secrets))
	keys := make([]string, 0, len(secrets))

	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		envvar, err := formatForEnv(key, secrets[key], secretsMap[key], tempFactory)
		if err != nil {
			return nil, err
		}
		env = append(env, envvar)
	}

	return env, nil
}

// resolveSecrets obtains the value of each requested secret.
func resolveSecrets(provider plugin_v1.Provider, secretsMap secretsyml.SecretsMap) (map[string]string, error) {
	var err error
	result := make(map[string]string)

	var varSecretsSpecKeys []string
	var varSecretsSpecPaths []string

	for key, spec := range secretsMap {
		if spec.IsVar() {
			varSecretsSpecKeys = append(varSecretsSpecKeys, key)
			varSecretsSpecPaths = append(varSecretsSpecPaths, spec.Path)
		} else {
			// If the spec isn't a variable, use its value as-is
			value := spec.Path
			result[key] = value
		}
	}

	// If there are no variables to resolve, return what we have
	if len(varSecretsSpecPaths) == 0 {
		return result, nil
	}

	// Get the variable values
	providerResponses, err := provider.GetValues(varSecretsSpecPaths...)
	if err != nil {
		return nil, err
	}

	// Collect errors from provider responses
	var errorStrings []string
	for _, providerResponse := range providerResponses {
		if providerResponse.Error != nil {
			errorStrings = append(errorStrings, providerResponse.Error.Error())
			continue
		}
	}
	if len(errorStrings) > 0 {
		// Sort the error strings to provide deterministic output
		sort.Strings(errorStrings)

		err = fmt.Errorf(strings.Join(errorStrings, "\n"))
		return nil, err
	}

	// Get variable values as strings
	for _, key := range varSecretsSpecKeys {
		spec := secretsMap[key]
		pr := providerResponses[spec.Path]
		result[key] = string(pr.Value)
	}

	return result, nil
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
func formatForEnv(
	key string,
	value string,
	spec secretsyml.SecretSpec,
	tempFactory *TempFactory,
) (string, error) {
	if spec.IsFile() {
		fname, err := tempFactory.Push(value)
		if err != nil {
			return "", err
		}

		value = fname
	}

	return fmt.Sprintf("%s=%s", key, value), nil
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

	if secrets, err = resolveSecrets(sc.Provider, sc.SecretsMap); err != nil {
		return
	}
	if env, err = buildEnvironment(secrets, sc.SecretsMap, sc.TempFactory); err != nil {
		return
	}

	return sc.runSubcommand(append(os.Environ(), env...))
}
