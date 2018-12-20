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
// TODO: turn to var and grab values from envvars for flexibility
const (
	TLS ServerTLSType = "DB_HOST_TLS"
	NoTLS = "DB_HOST_NO_TLS"
)
func ServerTLSTypeValues()[]ServerTLSType {
	return []ServerTLSType{TLS, NoTLS}
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
var (
	Undefined SSLRootCertType = ""
	Valid     SSLRootCertType = "Dynamically Set: Valid"
	Malformed SSLRootCertType = "malformed"
	Invalid   SSLRootCertType = "Dynamically Set: Invalid"
)

func SSLRootCertTypeValues()[]SSLRootCertType {
	return []SSLRootCertType{Undefined, Valid, Invalid, Malformed}
}

func (sslRootCertType SSLRootCertType) toConfigVariable() config.Variable {
	return config.Variable{
		Name:     "sslrootcert",
		Provider: "literal",
		ID:		   string(sslRootCertType),
	}
}
