// Code formatting/organization on this page is like this:
//
//     type def
//     type values
//     function to return all possible values
//     methods on the type
//
package test

import "github.com/cyberark/secretless-broker/pkg/secretless/config"

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

type ServerTLSSetting string
const (
	TLS   ServerTLSSetting = "DB_HOST_TLS"
	NoTLS                  = "DB_HOST_NO_TLS"
)
func AllServerTLSSettings()[]ServerTLSSetting {

	return []ServerTLSSetting{TLS, NoTLS}
}

func (tlsType ServerTLSSetting) toSecrets(dbConfig DBConfig) []config.StoredSecret {
	var secrets []config.StoredSecret
	var host string

	switch tlsType {
	case TLS:
		host = dbConfig.HostWithTLS
	case NoTLS:
		host = dbConfig.HostWithoutTLS
	default:
		panic("Invalid ServerTLSSetting provided")
	}

	switch dbConfig.Protocol {
	case "pg":
		secrets = append(secrets, config.StoredSecret{
			Name:     "address",
			Provider: "literal",
			ID:		  host + ":" + dbConfig.Port,
		})
	case "mysql":
		secrets = append(secrets, config.StoredSecret{
			Name:     "host",
			Provider: "literal",
			ID:		  host,
		})
		secrets = append(secrets, config.StoredSecret{
			Name:     "port",
			Provider: "literal",
			ID:		  dbConfig.Port,
		})
	default:
		panic("Invalid DB_PROTOCOL provided")
	}

	return secrets
}

type SSLModeType string
const (
	Default SSLModeType = ""
	Disable = "disable"
	Require = "require"
	VerifyCA = "verify-ca"
	VerifyFull = "verify-full"
)

func SSlModeTypeValues()[]SSLModeType {
	return []SSLModeType{Default, Disable, Require, VerifyCA, VerifyFull}
}

// For Secretless, sslmode="" is equivalent to not setting sslmode at all.
// Therefore, this will work for the "Default" case too.
func (sslMode SSLModeType) toConfigVariable() config.StoredSecret {
	return config.StoredSecret{
		Name:     "sslmode",
		Provider: "literal",
		ID:		   string(sslMode),
	}
}

type SSLRootCertType string

const (
	Undefined SSLRootCertType = ""
	Valid     SSLRootCertType = "/secretless/test/util/ssl/ca.pem"
	Malformed SSLRootCertType = "malformed"
	Invalid   SSLRootCertType = "/secretless/test/util/ssl/ca-invalid.pem"
)

func SSLRootCertTypeValues()[]SSLRootCertType {
	return []SSLRootCertType{Undefined, Valid, Invalid, Malformed}
}

func (sslRootCertType SSLRootCertType) toConfigVariable() config.StoredSecret {
	provider := "literal"
	switch sslRootCertType {
	case Valid, Invalid:
		provider = "file"
	}

	return config.StoredSecret{
		Name:     "sslrootcert",
		Provider: provider,
		ID:		   string(sslRootCertType),
	}
}

type SSLPrivateKeyType string
const (
	PrivateKeyUndefined SSLPrivateKeyType = ""
	PrivateKeyValid SSLPrivateKeyType = "/secretless/test/util/ssl/client-valid-key.pem"
	PrivateKeyNotSignedByCA SSLPrivateKeyType = "/secretless/test/util/ssl/client-different-ca-key.pem"
	PrivateKeyMalformed SSLPrivateKeyType = "malformed"
)

func SSLPrivateKeyTypeValues()[]SSLPrivateKeyType {
	return []SSLPrivateKeyType{PrivateKeyUndefined, PrivateKeyValid, PrivateKeyNotSignedByCA, PrivateKeyMalformed}
}

func (sslPrivateKeyType SSLPrivateKeyType) toConfigVariable() config.StoredSecret {
	provider := "literal"
	switch sslPrivateKeyType {
	case PrivateKeyValid, PrivateKeyNotSignedByCA:
		provider = "file"
	}

	return config.StoredSecret{
		Name:     "sslkey",
		Provider: provider,
		ID:		   string(sslPrivateKeyType),
	}
}

type SSLPublicCertType string
const (
	PublicCertUndefined SSLPublicCertType = ""
	PublicCertValid     SSLPublicCertType = "/secretless/test/util/ssl/client-valid.pem"
	PublicCertNotSignedByCA SSLPublicCertType = "/secretless/test/util/ssl/client-different-ca.pem"
	PublicCertMalformed SSLPublicCertType = "malformed"
)

func SSLPublicCertTypeValues()[]SSLPublicCertType {
	return []SSLPublicCertType{PublicCertUndefined, PublicCertValid, PublicCertNotSignedByCA, PublicCertMalformed}
}

func (sslPublicCertType SSLPublicCertType) toConfigVariable() config.StoredSecret {
	provider := "literal"
	switch sslPublicCertType {
	case PublicCertValid, PublicCertNotSignedByCA:
		provider = "file"
	}

	return config.StoredSecret{
		Name:     "sslcert",
		Provider: provider,
		ID:		   string(sslPublicCertType),
	}
}
