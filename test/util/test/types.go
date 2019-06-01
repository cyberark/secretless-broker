// Code formatting/organization on this page is like this:
//
//     type def
//     type values
//     function to return all possible values
//     methods on the type
//
package test

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

// NOTE: "Socket Type" is indeed the correct name here:
//       https://en.wikipedia.org/wiki/Network_socket#Other
type SocketType string
const (
	TCP    SocketType = "TCP"
	Socket            = "Unix Socket"
)
func AllSocketTypes()[]SocketType {
	return []SocketType{TCP, Socket}
}

type TLSSetting string

const (
	TLS   TLSSetting = "DB_HOST_TLS"
	NoTLS            = "DB_HOST_NO_TLS"
)
func AllTLSSettings()[]TLSSetting {
	return []TLSSetting{TLS, NoTLS}
}

//TODO: Something is still quite wrong with the design here:
//      Should pg/mysql logic live here?  It feels wrong...
func (tlsSetting TLSSetting) toSecrets(dbConfig DBConfig) []v1.StoredSecret {
	var secrets []v1.StoredSecret
	var host string

	switch tlsSetting {
	case TLS:
		host = dbConfig.HostWithTLS
	case NoTLS:
		host = dbConfig.HostWithoutTLS
	default:
		panic("Invalid TLSSetting")
	}

	switch dbConfig.Protocol {
	case "pg":
		secrets = append(secrets, v1.StoredSecret{
			Name:     "address",
			Provider: "literal",
			ID:		  host + ":" + dbConfig.Port,
		})
	case "mysql":
		secrets = append(secrets, v1.StoredSecret{
			Name:     "host",
			Provider: "literal",
			ID:		  host,
		})
		secrets = append(secrets, v1.StoredSecret{
			Name:     "port",
			Provider: "literal",
			ID:		  dbConfig.Port,
		})
	default:
		panic("Invalid DB_PROTOCOL provided")
	}

	return secrets
}

type SSLMode string
const (
	Default    SSLMode = ""
	Disable            = "disable"
	Require            = "require"
	VerifyCA           = "verify-ca"
	VerifyFull         = "verify-full"
)

func AllSSLModes()[]SSLMode {
	return []SSLMode{Default, Disable, Require, VerifyCA, VerifyFull}
}

// For Secretless, sslmode="" is equivalent to not setting sslmode at all.
// Therefore, this will work for the "Default" case too.
func (sslMode SSLMode) toSecret() v1.StoredSecret {
	return v1.StoredSecret{
		Name:     "sslmode",
		Provider: "literal",
		ID:		   string(sslMode),
	}
}

type RootCertStatus string

const (
	Undefined RootCertStatus = ""
	Valid                    = "/secretless/test/util/ssl/ca.pem"
	Malformed                = "malformed"
	Invalid                  = "/secretless/test/util/ssl/ca-invalid.pem"
)

func AllRootCertStatuses()[]RootCertStatus {
	return []RootCertStatus{Undefined, Valid, Invalid, Malformed}
}

func (sslRootCertType RootCertStatus) toSecret() v1.StoredSecret {
	provider := "literal"

	switch sslRootCertType {
	case Valid, Invalid:
		provider = "file"
	}

	return v1.StoredSecret{
		Name:     "sslrootcert",
		Provider: provider,
		ID:		  string(sslRootCertType),
	}
}

type PrivateKeyStatus string
const (
	PrivateKeyUndefined     PrivateKeyStatus = ""
	PrivateKeyValid                          = "/secretless/test/util/ssl/client-valid-key.pem"
	PrivateKeyNotSignedByCA                  = "/secretless/test/util/ssl/client-different-ca-key.pem"
	PrivateKeyMalformed                      = "malformed"
)

func AllPrivateKeyStatuses() []PrivateKeyStatus {
	return []PrivateKeyStatus{
		PrivateKeyUndefined, PrivateKeyValid, PrivateKeyNotSignedByCA, PrivateKeyMalformed,
	}
}

func (status PrivateKeyStatus) toSecret() v1.StoredSecret {

	provider := "literal"
	if status == PrivateKeyValid || status == PrivateKeyNotSignedByCA {
		provider = "file"
	}

	return v1.StoredSecret{
		Name:     "sslkey",
		Provider: provider,
		ID:       string(status),
	}
}

type PublicCertStatus string

const (
	PublicCertUndefined     PublicCertStatus = ""
	PublicCertValid                          = "/secretless/test/util/ssl/client-valid.pem"
	PublicCertNotSignedByCA                  = "/secretless/test/util/ssl/client-different-ca.pem"
	PublicCertMalformed                      = "malformed"
)

func AllPublicCertStatuses() []PublicCertStatus {
	return []PublicCertStatus{
		PublicCertUndefined, PublicCertValid, PublicCertNotSignedByCA, PublicCertMalformed,
	}
}

func (status PublicCertStatus) toSecret() v1.StoredSecret {

	provider := "literal"
	if status == PublicCertValid || status == PublicCertNotSignedByCA {
		provider = "file"
	}

	return v1.StoredSecret{
		Name:     "sslcert",
		Provider: provider,
		ID:       string(status),
	}
}
