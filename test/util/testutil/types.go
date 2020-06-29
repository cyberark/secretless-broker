// Package testutil has utilities for unit testing Secretless databases. Code
// formatting/organization on this page is like this:
//
//     type def
//     type values
//     function to return all possible values
//     methods on the type
//
// The design here is motivated by the desire to have a number of settings, each
// with a fixed set of values, so that we can loop through all combinations of
// them, to test all possibilities.
package testutil

import (
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// SocketType is either TCP or UNIX Socket.
// It is indeed the correct name here:
//     https://en.wikipedia.org/wiki/Network_socket#Other
type SocketType string

const (
	// TCP is a socket type
	TCP SocketType = "TCP"
	// Socket is a socket type
	Socket = "Unix Socket"
)

// AllSocketTypes returns the available socket types.
func AllSocketTypes() []SocketType {
	return []SocketType{TCP, Socket}
}

// TLSSetting can be TLS or NoTLS
type TLSSetting string

const (
	// TLS is a TLSSetting
	TLS TLSSetting = "DB_HOST_TLS"
	// NoTLS is a TLSSetting
	NoTLS = "DB_HOST_NO_TLS"
)

// AllTLSSettings returns the possible TLSSetting values: TLS or NoTLS
func AllTLSSettings() []TLSSetting {
	return []TLSSetting{TLS, NoTLS}
}

//TODO: Something is still quite wrong with the design here:
//      Should pg/mysql logic live here?  It feels wrong...
func (tlsSetting TLSSetting) toSecrets(dbConfig DBConfig) []*config_v2.Credential {
	var secrets []*config_v2.Credential
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
		secrets = append(secrets, &config_v2.Credential{
			Name: "host",
			From: "literal",
			Get:  host,
		})
		secrets = append(secrets, &config_v2.Credential{
			Name: "port",
			From: "literal",
			Get:  dbConfig.Port,
		})
	case "mysql":
		secrets = append(secrets, &config_v2.Credential{
			Name: "host",
			From: "literal",
			Get:  host,
		})
		secrets = append(secrets, &config_v2.Credential{
			Name: "port",
			From: "literal",
			Get:  dbConfig.Port,
		})
	default:
		panic("Invalid DB_PROTOCOL provided")
	}

	return secrets
}

// SSLMode describes possible SSL mode settings for a database.
type SSLMode string

const (
	// Default SSLMode
	Default SSLMode = ""
	// Disable SSLMode
	Disable = "disable"
	// Require SSLMode
	Require = "require"
	// VerifyCA SSLMode
	VerifyCA = "verify-ca"
	// VerifyFull SSLMode
	VerifyFull = "verify-full"
)

// AllSSLModes returns a list of all possible SSLMode values.
func AllSSLModes() []SSLMode {
	return []SSLMode{Default, Disable, Require, VerifyCA, VerifyFull}
}

// For Secretless, sslmode="" is equivalent to not setting sslmode at all.
// Therefore, this will work for the "Default" case too.
func (sslMode SSLMode) toSecret() *config_v2.Credential {
	return &config_v2.Credential{
		Name: "sslmode",
		From: "literal",
		Get:  string(sslMode),
	}
}

// SSLHost describes semantically possible sslhost settings when connecting to
// a database. sslhost specifies the value to carry out full verification
// against.
type SSLHost string

const (
	// SSLHostDefault is the default sslhost value which is empty
	SSLHostDefault SSLHost = ""
	// SSLHostInvalid is an invalid sslhost value
	SSLHostInvalid = "invalid"
)

// AllSSLHosts returns a list of all possible SSLHost values.
func AllSSLHosts() []SSLHost {
	return []SSLHost{SSLHostDefault, SSLHostInvalid}
}

// For Secretless, sslhost="" is equivalent to not setting sslhost at all.
// Therefore, this will work for the "Default" case too.
func (sslHost SSLHost) toSecret() *config_v2.Credential {
	return &config_v2.Credential{
		Name: "sslhost",
		From: "literal",
		Get:  string(sslHost),
	}
}

// AuthCredentialInvalidity specifies whether credentials are invalid.  We use
// Invalidity as opposed to CredentialValidity because bool defaults to false.
type AuthCredentialInvalidity bool

// AllAuthCredentialsInvalidity returns all possible values (which are just
// "true" and "false") that this setting can assume.
func AllAuthCredentialsInvalidity() []AuthCredentialInvalidity {
	return []AuthCredentialInvalidity{true, false}
}

func (authCredentialInvalidity AuthCredentialInvalidity) toSecrets() []*config_v2.Credential {
	password := sampleDbConfig.Password
	if authCredentialInvalidity {
		password = "wrong-password"
	}

	return []*config_v2.Credential{
		{
			Name: "username",
			From: "literal",
			Get:  sampleDbConfig.User,
		},
		{
			Name: "password",
			From: "literal",
			Get:  password,
		},
	}
}

// RootCertStatus represents possible statuses or states of the root cert.
type RootCertStatus string

const (
	// Undefined RootCertStatus
	Undefined RootCertStatus = ""
	// Valid RootCertStatus
	Valid = "/secretless/test/util/ssl/ca.pem"
	// Malformed RootCertStatus
	Malformed = "malformed"
	// Invalid RootCertStatus
	Invalid = "/secretless/test/util/ssl/ca-invalid.pem"
)

// AllRootCertStatuses returns all possible values for RootCertStatus.
func AllRootCertStatuses() []RootCertStatus {
	return []RootCertStatus{Undefined, Valid, Invalid, Malformed}
}

func (sslRootCertType RootCertStatus) toSecret() *config_v2.Credential {
	provider := "literal"

	switch sslRootCertType {
	case Valid, Invalid:
		provider = "file"
	}

	return &config_v2.Credential{
		Name: "sslrootcert",
		From: provider,
		Get:  string(sslRootCertType),
	}
}

// PrivateKeyStatus represents the status or state of a private key.
type PrivateKeyStatus string

const (
	// PrivateKeyUndefined PrivateKeyStatus
	PrivateKeyUndefined PrivateKeyStatus = ""
	// PrivateKeyValid PrivateKeyStatus
	PrivateKeyValid = "/secretless/test/util/ssl/client-key.pem"
	// PrivateKeyNotSignedByCA PrivateKeyStatus
	PrivateKeyNotSignedByCA = "/secretless/test/util/ssl/client-different-ca-key.pem"
	// PrivateKeyMalformed PrivateKeyStatus
	PrivateKeyMalformed = "malformed"
)

// AllPrivateKeyStatuses returns all possible values of PrivateKeyStatus.
func AllPrivateKeyStatuses() []PrivateKeyStatus {
	return []PrivateKeyStatus{
		PrivateKeyUndefined, PrivateKeyValid, PrivateKeyNotSignedByCA, PrivateKeyMalformed,
	}
}

func (status PrivateKeyStatus) toSecret() *config_v2.Credential {

	provider := "literal"
	if status == PrivateKeyValid || status == PrivateKeyNotSignedByCA {
		provider = "file"
	}

	return &config_v2.Credential{
		Name: "sslkey",
		From: provider,
		Get:  string(status),
	}
}

// PublicCertStatus represents the possible states of a public certificate.
type PublicCertStatus string

const (
	// PublicCertUndefined PublicCertStatus
	PublicCertUndefined PublicCertStatus = ""
	// PublicCertValid PublicCertStatus
	PublicCertValid = "/secretless/test/util/ssl/client.pem"
	// PublicCertNotSignedByCA PublicCertStatus
	PublicCertNotSignedByCA = "/secretless/test/util/ssl/client-different-ca.pem"
	// PublicCertMalformed PublicCertStatus
	PublicCertMalformed = "malformed"
)

// AllPublicCertStatuses returns all possible values for a PublicCertStatus
func AllPublicCertStatuses() []PublicCertStatus {
	return []PublicCertStatus{
		PublicCertUndefined, PublicCertValid, PublicCertNotSignedByCA, PublicCertMalformed,
	}
}

func (status PublicCertStatus) toSecret() *config_v2.Credential {

	provider := "literal"
	if status == PublicCertValid || status == PublicCertNotSignedByCA {
		provider = "file"
	}

	return &config_v2.Credential{
		Name: "sslcert",
		From: provider,
		Get:  string(status),
	}
}
