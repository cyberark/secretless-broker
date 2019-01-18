package tests

import (
	. "github.com/cyberark/secretless-broker/test/util/test"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSSL(t *testing.T) {

	testCases := []TestCase{
		{
			TestDefinition: TestDefinition{
				Description:   "server_tls, sslmode=default",
				ShouldPass: true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Default,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:   "server_tls, sslmode=disable",
				ShouldPass: true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Disable,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:   "server_tls, sslmode=require, sslrootcert=none",
				ShouldPass: true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Require,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=require, sslrootcert=invalid (ignored)",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Require,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=require, sslrootcert=malformed (ignored)",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Require,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=none",
				ShouldPass:  false,
				CmdOutput:   StringPointer("ERROR 2000 (HY000): x509: certificate signed by unknown authority"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        VerifyCA,
				RootCertStatus: Undefined,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:   "server_tls, sslmode=verify-ca, sslrootcert=valid",
				ShouldPass:    true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        VerifyCA,
				RootCertStatus: Valid,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=invalid",
				ShouldPass:  false,
				CmdOutput:   StringPointer(`ERROR 2000 (HY000): x509: certificate signed by unknown authority`),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        VerifyCA,
				RootCertStatus: Invalid,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=malformed",
				ShouldPass:  false,
				CmdOutput:   StringPointer("ERROR 2000 (HY000): couldn't parse pem in sslrootcert"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        VerifyCA,
				RootCertStatus: Malformed,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_no_tls, sslmode=default",
				ShouldPass:  false,
				CmdOutput:   StringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     NoTLS,
				SSLMode:        Default,
				RootCertStatus: Undefined,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:   "server_no_tls, sslmode=disable",
				ShouldPass:    true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     NoTLS,
				SSLMode:        Disable,
				RootCertStatus: Undefined,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:         "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=malformed, sslcert=malformed",
				ShouldPass:          false,
				CmdOutput:           StringPointer("ERROR 2000 (HY000): tls: failed to find any PEM data in certificate input"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:       TCP,
				TLSSetting:       TLS,
				SSLMode:          VerifyCA,
				RootCertStatus:   Valid,
				PrivateKeyStatus: PrivateKeyMalformed,
				PublicCertStatus: PublicCertMalformed,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:         "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=valid, sslcert=valid",
				ShouldPass:          true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:       TCP,
				TLSSetting:       TLS,
				SSLMode:          VerifyCA,
				RootCertStatus:   Valid,
				PrivateKeyStatus: PrivateKeyValid,
				PublicCertStatus: PublicCertValid,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description:         "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=notsignedbyca, sslcert=notsignedbyca",
				ShouldPass:          false,
				CmdOutput:           StringPointer("ERROR 2000 (HY000): remote error: tls: unknown certificate authority"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:       TCP,
				TLSSetting:       TLS,
				SSLMode:          VerifyCA,
				RootCertStatus:   Valid,
				PrivateKeyStatus: PrivateKeyNotSignedByCA,
				PublicCertStatus: PublicCertNotSignedByCA,
			},
		},
	}

	Convey("SSL functionality", t, func() {
		for _, testCase := range testCases {
			RunTestCase(testCase)
		}
	})
}
