package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/conjurinc/secretless/internal/app/summon/command"

	. "github.com/smartystreets/goconvey/convey"
	"os"
	"io"
)

// TestSummon2_Run tests Summon at the CLI level, including argument parsing etc.
func TestSummon2_Cmd(t *testing.T) {
	var err error

	Convey("Provides secrets to a subprocess environment", t, func() {
		secretsDescriptor := `
DB_PASSWORD: literal-password
`
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		args := []string{"summon2", "--yaml", secretsDescriptor, "env"}
		err = command.RunCLI(args, nil)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()

		output := string(buf.Bytes())
		lines := strings.Split(output, "\n")

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=literal-password")
	})
}
