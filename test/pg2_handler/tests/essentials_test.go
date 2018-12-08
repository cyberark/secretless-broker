package tests

import (
	"github.com/cyberark/secretless-broker/test/pg2_handler/pkg"
	. "github.com/cyberark/secretless-broker/test/util/test"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEssentials(t *testing.T) {

	// TODO: figure out how to handle the naming convention of unix domain sockets for pg
	// see https://www.postgresql.org/docs/9.3/runtime-config-connection.html#GUC-UNIX-SOCKET-DIRECTORIES
	// perhaps have test local package pass a NameGenerator that takes the port number
	// we'll need to get the port number from the socket file and create the appropriate flag
	convey.Convey("Connect over TCP", t, func() {
		tc := TestCase{
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
			TestDefinition: TestDefinition{
				Description: "with username, wrong password",
				Flags: pkg.ConnectionParams{
					Username: "testuser",
					Password: "wrongpassword",
					SSLMode:  "disable",
				}.ToFlags(),
				ShouldPass: true,
				CmdOutput:  StringPointer(`1 row`),
			},
		}

		RunTestCase(tc)
	})
}
