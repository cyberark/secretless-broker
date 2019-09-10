package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/secretless-broker/internal/summon/command"
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

	Convey("Provides secrets to a subprocess environment", t, func() {
		args := []string{"summon2", "-p", "literal", "--yaml", secretsDescriptor, "env"}

		lines, err := runCommand(args)

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=literal-password")
	})

	Convey("Provider can be specified as an environment variable", t, func() {
		err := os.Setenv("SUMMON_PROVIDER", "literal")
		So(err, ShouldBeNil)
		defer os.Unsetenv("SUMMON_PROVIDER")

		lines, err := runCommand(defaultArgs)

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=literal-password")
	})
}
