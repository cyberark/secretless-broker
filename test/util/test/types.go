// Code formatting/organization on this page is like this:
//
//     type def
//     type values
//     function to return all possible values
//     methods on the type
//
package test

import "github.com/cyberark/secretless-broker/pkg/secretless/config"

type ListenerType string
const (
	TCP ListenerType = "TCP"
	Socket = "Unix Socket"
)
func ListenerTypeValues()[]ListenerType {
	return []ListenerType{TCP, Socket}
}

type ServerTLSType string
const (
	TLS ServerTLSType = "DB_HOST_TLS"
	NoTLS = "DB_HOST_NO_TLS"
)
func ServerTLSTypeValues()[]ServerTLSType {
	return []ServerTLSType{TLS, NoTLS}
}

func (tlsType ServerTLSType) toConfigVariables(dbConfig TestDBConfigType) []config.Variable  {
	variables := []config.Variable{}
	var host string
	switch tlsType {
	case TLS:
		host = dbConfig.DB_HOST_TLS
	case NoTLS:
		host = dbConfig.DB_HOST_NO_TLS
	default:
		panic("Invalid ServerTLSType provided")
	}

	switch dbConfig.DB_PROTOCOL {
	case "pg":
		variables = append(variables, config.Variable{
			Name:     "address",
			Provider: "literal",
			ID:		  host + ":" + dbConfig.DB_PORT,
		})
	case "mysql":
		variables = append(variables, config.Variable{
			Name:     "host",
			Provider: "literal",
			ID:		  host,
		})
		variables = append(variables, config.Variable{
			Name:     "port",
			Provider: "literal",
			ID:		  dbConfig.DB_PORT,
		})
	default:
		panic("Invalid DB_PROTOCOL provided")
	}

	return variables
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
func (sslMode SSLModeType) toConfigVariable() config.Variable {
	return config.Variable{
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

func (sslRootCertType SSLRootCertType) toConfigVariable() config.Variable {
	provider := "literal"
	switch sslRootCertType {
	case Valid, Invalid:
		provider = "file"
	}

	return config.Variable{
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

func (sslPrivateKeyType SSLPrivateKeyType) toConfigVariable() config.Variable {
	provider := "literal"
	switch sslPrivateKeyType {
	case PrivateKeyValid, PrivateKeyNotSignedByCA:
		provider = "file"
	}

	return config.Variable{
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

func (sslPublicCertType SSLPublicCertType) toConfigVariable() config.Variable {
	provider := "literal"
	switch sslPublicCertType {
	case PublicCertValid, PublicCertNotSignedByCA:
		provider = "file"
	}

	return config.Variable{
		Name:     "sslcert",
		Provider: provider,
		ID:		   string(sslPublicCertType),
	}
}
