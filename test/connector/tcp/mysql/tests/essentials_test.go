package tests

import (
	"fmt"
	"testing"

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
			},
		},
		{
			Description: "with wrong username, wrong password",
			ShouldPass:  true,
			ClientConfiguration: ClientConfiguration{
				Username: "wrongusername",
				Password: "wrongpassword",
			},
		},
		{
			Description: "with empty username, empty password",
			ShouldPass:  true,
			ClientConfiguration: ClientConfiguration{
				Username: "",
				Password: "",
			},
		},
	}

	t.Run("Essentials", func(t *testing.T) {
		for _, socketType := range AllSocketTypes() {
			t.Run(fmt.Sprintf("Connect over %s", socketType), func(t *testing.T) {

				for _, testCaseData := range testCases {
					tc := TestCase{
						AbstractConfiguration: AbstractConfiguration{
							SocketType:     socketType,
							TLSSetting:     TLS,
							SSLMode:        Default,
							RootCertStatus: Undefined,
						},
						Definition: testCaseData,
					}
					RunTestCase(tc, t)
				}
			})
		}

		// TODO: check client net.conn for mysql and postgres
		// if connected via socket then there's no need to check if the client wants TLS
		// assume no TLS between client and secretless
		// NOTE: this is the default behaviour of psql not mysql
		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     Socket,
				TLSSetting:     TLS,
				SSLMode:        Default,
				RootCertStatus: Undefined,
			},
			Definition: Definition{
				Description: "Socket, client -> TLS -> secretless",
				ShouldPass:  false,
				ClientConfiguration: ClientConfiguration{
					Username: "wrongusername",
					Password: "wrongpassword",
					SSL:      true,
				},
				CmdOutput: StringPointer("ERROR 2026 (HY000): TLS/SSL error: SSL is required, but the server does not support it"),
			},
		}, t)

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
				CmdOutput: StringPointer("ERROR 2026 (HY000): TLS/SSL error: SSL is required, but the server does not support it"),
			},
		}, t)

		RunTestCase(TestCase{
			AbstractConfiguration: AbstractConfiguration{
				SocketType:               TCP,
				TLSSetting:               TLS,
				SSLMode:                  Default,
				RootCertStatus:           Undefined,
				AuthCredentialInvalidity: true,
			},
			Definition: Definition{
				Description: "secretless using invalid credentials",
				ShouldPass:  false,
				ClientConfiguration: ClientConfiguration{
					Username: "testuser",
					Password: "wrongpassword",
				},
				CmdOutput: StringPointer("ERROR 1045 (28000): Access denied for user 'testuser'@"),
			},
		}, t)
	})

}
