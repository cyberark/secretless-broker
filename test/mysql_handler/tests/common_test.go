package tests

import (
	"fmt"
	"testing"

	. "github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCommonMySQLHandler(t *testing.T) {
	testCases := []TestData{
		{
			Description: "with username, wrong password",
			Flags: []string{
				"--user=testuser",
				"--password=wrongpassword",
			},
		},
		{
			Description: "with wrong username, wrong password",
			Flags: []string{
				"--user=wrongusername",
				"--password=wrongpassword",
			},
		},
		{
			Description: "with empty username, empty password",
			Flags: []string{
				"--user=",
				"--password=",
			},
		},
		{
			Description: "client -> TLS -> secretless",
			Flags: []string{
				"--user=",
				"--password=",
				"--ssl-verify-server-cert=TRUE",
				"--ssl",
			},
			AssertFailure: true,
			CmdOutput: StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required, but the server does not support"),
		},
	}

	for _, listenerTypeValue := range ListenerTypeValues {
		Convey(fmt.Sprintf("Connect over %s", listenerTypeValue), t, func() {

			for _, testCaseData := range testCases {
				tc := TestCase{
					AbstractConfiguration: AbstractConfiguration{
						ListenerType:    listenerTypeValue,
						ServerTLSType:   TLS,
						SSLModeType:     Default,
						SSLRootCertType: Undefined,
					},
					TestData: testCaseData,
				}
				Runner(tc)
			}
		})
	}

}
