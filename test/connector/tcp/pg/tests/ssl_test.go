package tests

import (
	"testing"

	. "github.com/cyberark/secretless-broker/test/util/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSSL(t *testing.T) {

	testCases := []TestCase{
		{
			Definition: Definition{
				Description: "server_tls, sslmode=default",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Default,
			},
		},
		{
			Definition: Definition{
				Description: "server_tls, sslmode=disable",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Disable,
			},
		},
		{
			Definition: Definition{
				Description: "server_tls, sslmode=require, sslrootcert=none",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType: TCP,
				TLSSetting: TLS,
				SSLMode:    Require,
			},
		},
		{
			Definition: Definition{
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
			Definition: Definition{
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
			Definition: Definition{
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
			Definition: Definition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=valid",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     TLS,
				SSLMode:        VerifyCA,
				RootCertStatus: Valid,
			},
		},
		{
			Definition: Definition{
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
			Definition: Definition{
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
			Definition: Definition{
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
			Definition: Definition{
				Description: "server_no_tls, sslmode=disable",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:     TCP,
				TLSSetting:     NoTLS,
				SSLMode:        Disable,
				RootCertStatus: Undefined,
			},
		},
		{
			Definition: Definition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=malformed, sslcert=malformed",
				ShouldPass:  false,
				CmdOutput:   StringPointer("psql: FATAL:  tls: failed to find any PEM data in certificate input"),
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
			Definition: Definition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=valid, sslcert=valid",
				ShouldPass:  true,
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
			Definition: Definition{
				Description: "server_tls, sslmode=verify-ca, sslrootcert=valid, sslkey=notsignedbyca, sslcert=notsignedbyca",
				ShouldPass:  true,
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
		{
			Definition: Definition{
				Description: "server_tls, sslmode=verify-full, sslrootcert=valid, sslkey=valid, sslcert=valid, sslhost=valid",
				ShouldPass:  true,
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:       TCP,
				TLSSetting:       TLS,
				SSLMode:          VerifyFull,
				RootCertStatus:   Valid,
				PrivateKeyStatus: PrivateKeyValid,
				PublicCertStatus: PublicCertValid,
			},
		},
		{
			Definition: Definition{
				Description: "server_tls, sslmode=verify-full, sslrootcert=valid, sslkey=valid, sslcert=valid, sslhost=invalid",
				ShouldPass:  false,
				CmdOutput:   StringPointer("psql: FATAL:  x509: certificate is valid for localhost, mysql, pg, not invalid"),
			},
			AbstractConfiguration: AbstractConfiguration{
				SocketType:       TCP,
				TLSSetting:       TLS,
				SSLMode:          VerifyFull,
				SSLHost:          SSLHostInvalid,
				RootCertStatus:   Valid,
				PrivateKeyStatus: PrivateKeyValid,
				PublicCertStatus: PublicCertValid,
			},
		},
	}

	Convey("SSL functionality", t, func() {
		for _, testCase := range testCases {
			RunTestCase(testCase)
		}
	})
}
