package pkg

type ListenerType string
const (
	TCP ListenerType = "TCP"
	Socket = "Unix Socket"
)
var ListenerTypeValues = []ListenerType{TCP, Socket}


type ServerTLSType string
// TODO: turn to var and grab values from envvars for flexibility
const (
	TLS ServerTLSType = "mysql"
	NoTLS = "mysql_no_tls"
)
var ServerTLSTypeValues = []ServerTLSType{TLS, NoTLS}


type SSLModeType string
const (
	Default SSLModeType = ""
	Disable = "disable"
	Require = "require"
	VerifyCA = "verify-ca"
	VerifyFull = "verify-full"
)
var SSlModeTypeValues = []SSLModeType{Default, Disable, Require, VerifyCA, VerifyFull}


type SSLRootCertType string
// TODO: turn to var and grab values from envvars for flexibility
const (
	Undefined SSLRootCertType = ""
	Valid = "TODO: add valid"
	Invalid = "TODO: add invalid"
)
var SSLRootCertTypeValue = []SSLRootCertType{Undefined, Valid, Invalid}
