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
			Flags: []string{
				"--user=testuser",
				"--password=wrongpassword",
			},
		},
		{
			Description: "with wrong username, wrong password",
			ShouldPass: true,
			Flags: []string{
				"--user=wrongusername",
				"--password=wrongpassword",
			},
		},
		{
			Description: "with empty username, empty password",
			ShouldPass: true,
			Flags: []string{
				"--user=",
				"--password=",
			},
		},
		{
			Description: "client -> TLS -> secretless",
			ShouldPass: false,
			Flags: []string{
				"--user=",
				"--password=",
				"--ssl-verify-server-cert=TRUE",
				"--ssl",
			},
			CmdOutput:  StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required, but the server does not support"),
		},
	}

	for _, listenerTypeValue := range ListenerTypeValues() {
		Convey(fmt.Sprintf("Connect over %s", listenerTypeValue), t, func() {

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

}
