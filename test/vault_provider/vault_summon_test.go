package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/cyberark/secretless-broker/internal/app/summon/command"
	. "github.com/smartystreets/goconvey/convey"

	_ "github.com/joho/godotenv/autoload"
)

func TestVault_Summon(t *testing.T) {
	secretsDescriptor := `
DB_PASSWORD: !var kv/db/password#password
WEB_PASSWORD: !var kv/web/password
`
	defaultArgs := []string{"summon2", "-p", "vault", "--yaml", secretsDescriptor, "env"}

	runCommand := func(args []string) (lines []string, err error) {
		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)

		err = command.RunCLI(args, writer)

		writer.Flush()
		output := string(buffer.Bytes())
		lines = strings.Split(output, "\n")

		return
	}

	Convey("Can summon secrets from Vault", t, func() {
		lines, err := runCommand(defaultArgs)

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=db-secret")
		So(lines, ShouldContain, "WEB_PASSWORD=web-secret")
	})
}
