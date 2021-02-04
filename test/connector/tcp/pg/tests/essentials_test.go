package tests

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/secretless-broker/test/connector/tcp/pg/pkg"
	. "github.com/cyberark/secretless-broker/test/util/testutil"
)

func TestEssentials(t *testing.T) {
	testCases := []Definition{
		{
			Description: "with username, wrong password",
			ShouldPass:  true,
			ClientConfiguration: ClientConfiguration{
				Username: "testuser",
				Password: "wrongpassword",
				SSL:      false,
			},
		},
		{
			Description: "with wrong username, wrong password",
			ShouldPass:  true,
			ClientConfiguration: ClientConfiguration{
				Username: "wrongusername",
				Password: "wrongpassword",
				SSL:      false,
			},
		},
		{
			Description: "with empty username, empty password",
			ShouldPass:  true,
			ClientConfiguration: ClientConfiguration{
				Username: "",
				Password: "",
				SSL:      false,
			},
		},
	}

	Convey("Essentials", t, func() {
		for _, listenerTypeValue := range AllSocketTypes() {
			Convey(fmt.Sprintf("Connect over %s", listenerTypeValue), func() {

				for _, testCaseData := range testCases {
					tc := TestCase{
						AbstractConfiguration: AbstractConfiguration{
							SocketType:     listenerTypeValue,
							TLSSetting:     TLS,
							SSLMode:        Default,
							RootCertStatus: Undefined,
						},
						Definition: testCaseData,
					}
					RunTestCase(tc)
				}
			})
		}

		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     Socket,
				TLSSetting:     TLS,
				SSLMode:        Default,
				RootCertStatus: Undefined,
			},
			Definition: Definition{
				Description: "Socket, client -> TLS -> secretless",
				ShouldPass:  true,
				ClientConfiguration: ClientConfiguration{
					Username: "wrongusername",
					Password: "wrongpassword",
					SSL:      true,
				},
			},
		})

		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        Default,
				RootCertStatus: Undefined,
			},
			Definition: Definition{
				Description: "TCP, client -> TLS -> secretless",
				ShouldPass:  false,
				ClientConfiguration: ClientConfiguration{
					Username: "wrongusername",
					Password: "wrongpassword",
					SSL:      true,
				},
				CmdOutput: StringPointer("server does not support SSL, but SSL was required"),
			},
		})
	})

	Convey("JDBC", t, func() {
		RunJDBCTestCase := NewRunTestCase(pkg.RunJDBCQuery)

		Convey(fmt.Sprintf("Connect over %s", TCP), func() {

			for _, testCaseData := range testCases {
				tc := TestCase{
					AbstractConfiguration: AbstractConfiguration{
						SocketType:     TCP,
						TLSSetting:     TLS,
						SSLMode:        Default,
						RootCertStatus: Undefined,
					},
					Definition: testCaseData,
				}
				RunJDBCTestCase(tc)
			}
		})
	})
}
