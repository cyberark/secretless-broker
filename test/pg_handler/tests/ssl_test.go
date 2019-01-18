package tests

import (
	"testing"

	. "github.com/cyberark/secretless-broker/test/util/test"
	. "github.com/smartystreets/goconvey/convey"
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
				Description: "server_tls, sslmode=require, sslrootcert=invalid",
				ShouldPass:  false,
				CmdOutput:   StringPointer(`x509: certificate signed by unknown authority`),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        Require,
				RootCertStatus: Invalid,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=require, sslrootcert=malformed",
				ShouldPass:  false,
				CmdOutput:   StringPointer("couldn't parse pem in sslrootcert"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        Require,
				RootCertStatus: Malformed,
			},
		},
		{
			TestDefinition: TestDefinition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=none",
				ShouldPass:  false,
				CmdOutput:   StringPointer("x509: certificate signed by unknown authority"),
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
				CmdOutput:   StringPointer(`certificate signed by unknown authority`),
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
				CmdOutput:   StringPointer("couldn't parse pem in sslrootcert"),
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
				CmdOutput:   StringPointer("the backend does not allow SSL connections"),
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
				CmdOutput:           StringPointer("psql: FATAL:  tls: failed to find any PEM data in certificate input"),
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
				ShouldPass:          true,
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
