package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cyberark/secretless-broker/internal/summon/command"
	"github.com/stretchr/testify/assert"
)

// TestSummon2_Run tests Summon at the CLI level, including argument parsing etc.
func TestSummon2_Cmd(t *testing.T) {
	secretsDescriptor := `
DB_PASSWORD: literal-password
`
	defaultArgs := []string{"summon2", "--yaml", secretsDescriptor, "env"}

	runCommand := func(args []string) (lines []string, err error) {
		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)

		err = command.RunCLI(args, writer)

		writer.Flush()
		output := string(buffer.Bytes())
		lines = strings.Split(output, "\n")

		return
	}

	t.Run("Provides secrets to a subprocess environment", func(t *testing.T) {
		args := []string{"summon2", "-p", "literal", "--yaml", secretsDescriptor, "env"}

		lines, err := runCommand(args)

		assert.NoError(t, err)
		assert.Contains(t, lines, "DB_PASSWORD=literal-password")
	})

	t.Run("Provider can be specified as an environment variable", func(t *testing.T) {
		err := os.Setenv("SUMMON_PROVIDER", "literal")
		assert.NoError(t, err)
		defer os.Unsetenv("SUMMON_PROVIDER")

		lines, err := runCommand(defaultArgs)

		assert.NoError(t, err)
		assert.Contains(t, lines, "DB_PASSWORD=literal-password")
	})
}
