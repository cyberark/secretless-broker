package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/conjurinc/secretless/internal/app/summon/command"

	. "github.com/smartystreets/goconvey/convey"
	"bufio"
)

// TestSummon2_Run tests Summon at the CLI level, including argument parsing etc.
func TestSummon2_Cmd(t *testing.T) {
	var err error

	Convey("Provides secrets to a subprocess environment", t, func() {
		secretsDescriptor := `
DB_PASSWORD: literal-password
`
		args := []string{"summon2", "--yaml", secretsDescriptor, "env"}

		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)

		err = command.RunCLI(args, writer)

		writer.Flush()
		output := string(buffer.Bytes())
		lines := strings.Split(output, "\n")

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=literal-password")
	})
}
