package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/secretless-broker/internal/summon/command"
)

func TestVault_Summon(t *testing.T) {
	secretsDescriptor := `
FIRST_SECRET: !var cubbyhole/first-secret#some-key
SECOND_SECRET: !var cubbyhole/second-secret
DB_PASSWORD: !var kv/db/password#password
WEB_PASSWORD: !var kv/web/password
SVC_API_KEY: !var secret/data/service#data.api-key
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
		So(lines, ShouldContain, "FIRST_SECRET=one")
		So(lines, ShouldContain, "SECOND_SECRET=two")
		So(lines, ShouldContain, "DB_PASSWORD=db-secret")
		So(lines, ShouldContain, "WEB_PASSWORD=web-secret")
		So(lines, ShouldContain, "SVC_API_KEY=service-api-key")
	})
}
