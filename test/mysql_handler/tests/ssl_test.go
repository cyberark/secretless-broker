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
				Description:   "server_tls, sslmode=default",
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Default,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=disable",
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Disable,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=require, sslrootcert=none",
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Require,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=require, sslrootcert=invalid",
				AssertFailure: true,
				CmdOutput:     StringPointer(`ERROR 2000 (HY000): x509: certificate signed by unknown authority`),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Require,
				SSLRootCertType: Invalid,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=require, sslrootcert=malformed",
				AssertFailure: true,
				CmdOutput:     StringPointer("ERROR 2000 (HY000): couldn't parse pem in sslrootcert"),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     Require,
				SSLRootCertType: Malformed,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=verify-ca, sslrootcert=none",
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
				Description:   "server_tls, sslmode=verify-ca, sslrootcert=valid",
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     VerifyCA,
				SSLRootCertType: Valid,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=verify-ca, sslrootcert=invalid",
				AssertFailure: true,
				CmdOutput:     StringPointer(`ERROR 2000 (HY000): x509: certificate signed by unknown authority`),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     VerifyCA,
				SSLRootCertType: Invalid,
			},
		},
		{
			TestData: TestData{
				Description:   "server_tls, sslmode=verify-ca, sslrootcert=malformed",
				AssertFailure: true,
				CmdOutput:     StringPointer("ERROR 2000 (HY000): couldn't parse pem in sslrootcert"),
			},
			AbstractConfiguration: AbstractConfiguration{
				ListenerType:    TCP,
				ServerTLSType:   TLS,
				SSLModeType:     VerifyCA,
				SSLRootCertType: Malformed,
			},
		},
		{
			TestData: TestData{
				Description:   "server_no_tls, sslmode=default",
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
