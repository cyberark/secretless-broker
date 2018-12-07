package pkg

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
	TLS ServerTLSType = "mysql"
	NoTLS = "mysql_no_tls"
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

func (sslMode SSLModeType) toConfigVariable() config.Variable {
	return config.Variable{
		Name:     "sslmode",
		Provider: "literal",
		ID:		   string(sslMode),
	}
}

type SSLRootCertType string
// TODO: turn to var and grab values from envvars for flexibility
var (
	Undefined SSLRootCertType = ""
	Valid     SSLRootCertType = "TODO: add valid"
	Malformed SSLRootCertType = "malformed"
	Invalid   SSLRootCertType = "TODO: add invalid"
)
func SSLRootCertTypeValues()[]SSLRootCertType {
	return []SSLRootCertType{Undefined, Valid, Invalid, Malformed}
}
