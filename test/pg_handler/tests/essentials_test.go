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
				Username: StringPointer("testuser"),
				Password: StringPointer("wrongpassword"),
			},
		},
		{
			Description: "with wrong username, wrong password",
			ShouldPass: true,
			ClientConfiguration: ClientConfiguration{
				Username: StringPointer("wrongusername"),
				Password: StringPointer("wrongpassword"),
			},
		},
		{
			Description: "with empty username, empty password",
			ShouldPass: true,
			ClientConfiguration: ClientConfiguration{
				Username: StringPointer(""),
				Password: StringPointer(""),
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

		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    Socket,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
			TestDefinition: TestDefinition{
				Description: "Socket, client -> TLS -> secretless",
				ShouldPass:  true,
				ClientConfiguration: ClientConfiguration{
					Username: StringPointer("wrongusername"),
					Password: StringPointer("wrongpassword"),
					SSL:      BoolPointer(true),
				},
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
					Username: StringPointer("wrongusername"),
					Password: StringPointer("wrongpassword"),
					SSL:      BoolPointer(true),
				},
				CmdOutput: StringPointer("SSL not supported"),
			},
		})
	})

}
