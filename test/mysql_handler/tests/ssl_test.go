package tests

import (
	"testing"

	. "github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSSLMySQLHandler(t *testing.T) {

	testCases := []TestCase{
		{
			TestData: TestData{
				Description:   "server_tls",
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=verify-ca",
				AssertFailure: true,
				CmdOutput:     StringPointer("ERROR 2000 (HY000): x509: certificate signed by unknown authority"),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     VerifyCA,
				SSLRootCertType: Undefined,
			},
		},
		{
			TestData: TestData{
				Description:   "server_no_tls",
				AssertFailure: true,
				CmdOutput:     StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it"),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   NoTLS,
				SSLModeType:     Default,
				SSLRootCertType: Undefined,
			},
		},
		{
			TestData: TestData{
				Description:   "server_no_tls, sslmode=disable",
				Flags:         nil,
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   NoTLS,
				SSLModeType:     Disable,
				SSLRootCertType: Undefined,
			},
		},
	}

	Convey("SSL functionality", t, func() {
		for _, testCase := range testCases {
			Runner(testCase)
		}
	})
}
