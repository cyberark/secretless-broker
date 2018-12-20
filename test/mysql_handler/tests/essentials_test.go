package tests

import (
	"fmt"
	"testing"

	. "github.com/cyberark/secretless-broker/test/util/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEssentials(t *testing.T) {
	testCases := []TestDefinition{
		{
			Description: "with username, wrong password",
			ShouldPass: true,
			ClientConfiguration: ClientConfiguration{
				Username: "testuser",
				Password: "wrongpassword",
			},
		},
		{
			Description: "with wrong username, wrong password",
			ShouldPass: true,
			ClientConfiguration: ClientConfiguration{
				Username: "wrongusername",
				Password: "wrongpassword",
			},
		},
		{
			Description: "with empty username, empty password",
			ShouldPass: true,
			ClientConfiguration: ClientConfiguration{
				Username: "",
				Password: "",
			},
		},
	}

	Convey("Essentials", t, func() {
		for _, listenerTypeValue := range ListenerTypeValues() {
			Convey(fmt.Sprintf("Connect over %s", listenerTypeValue), func() {

				for _, testCaseData := range testCases {
					tc := TestCase{
						AbstractConfiguration: AbstractConfiguration{
							ListenerType:    listenerTypeValue,
							ServerTLSType:   TLS,
							SSLModeType:     Default,
							SSLRootCertType: Undefined,
						},
						TestDefinition: testCaseData,
					}
					RunTestCase(tc)
				}
			})
		}

		// TODO: check client net.conn for mysql and postgres
		// if connected via socket then there's no need to check if the client wants TLS
		// assume no TLS between client and secretless
		// NOTE: this is the default behaviour of psql not mysql
		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    Socket,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
			TestDefinition: TestDefinition{
				Description: "Socket, client -> TLS -> secretless",
				ShouldPass:  false,
				ClientConfiguration: ClientConfiguration{
					Username: "wrongusername",
					Password: "wrongpassword",
					SSL:      true,
				},
				CmdOutput: StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required, but the server does not support"),
			},
		})

		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
			TestDefinition: TestDefinition{
				Description: "TCP, client -> TLS -> secretless",
				ShouldPass:  false,
				ClientConfiguration: ClientConfiguration{
					Username: "wrongusername",
					Password: "wrongpassword",
					SSL:      true,
				},
				CmdOutput: StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required, but the server does not support"),
			},
		})
	})

}
