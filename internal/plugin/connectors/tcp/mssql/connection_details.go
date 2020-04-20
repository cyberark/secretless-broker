package mssql

import (
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/connectiondetails"
	"net/url"
)

var sslModeToBaseParams = map[string]map[string]string{
	sslModeDisable: {
		"encrypt": "disable",
	},
	sslModeRequire: {
		"encrypt":                "true",
		"trustservercertificate": "true",
	},
	sslModeVerifyCA: {
		"encrypt":                "true",
		"trustservercertificate": "false",
		"disableverifyhostname":  "true",
	},
	sslModeVerifyFull: {
		"encrypt":                "true",
		"trustservercertificate": "false",
		"disableverifyhostname":  "false",
	},
}

const (
	sslModeDisable    = "disable"
	sslModeRequire    = "require"
	sslModeVerifyCA   = "verify-ca"
	sslModeVerifyFull = "verify-full"
)

var defaultSSLMode = []byte(sslModeRequire)

const defaultMSSQLPort = "1433"

// NewConnectionDetails is a local constructor for creating connection details and
// injecting custom handling for MsSQL-specific parameters
func NewConnectionDetails(credentials map[string][]byte) (*connectiondetails.ConnectionDetails, error) {
	return connectiondetails.NewConnectionDetails(
		credentials,
		defaultMSSQLPort,
		HandleSSLOptions,
	)
}

// HandleSSLOptions is a custom handler for MsSQL, converting the sslmode to the
// corresponding value that is understood by MsSQL, and adding any needed params.
func HandleSSLOptions(credentials map[string][]byte) (map[string]string, error) {
	sslMode := string(credentials["sslmode"])
	params, ok := sslModeToBaseParams[sslMode]

	if !ok {
		credentials["sslmode"] = defaultSSLMode
		return HandleSSLOptions(credentials)
	}

	if sslMode == sslModeVerifyCA {
		params["rawcertificate"] = string(credentials["sslrootcert"])
	}

	if sslMode == sslModeVerifyFull {
		params["rawcertificate"] = string(credentials["sslrootcert"])

		// Ability to override hostname for verification
		if len(credentials["sslhost"]) > 0 {
			params["hostnameincertificate"] = string(credentials["sslhost"])
		}
	}

	return params, nil
}

// URL returns a string URL from connection details
func URL(cd *connectiondetails.ConnectionDetails) string {
	query := url.Values{}

	for key, value := range cd.Options {
		query.Add(key, value)
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cd.Username, cd.Password),
		Host:     cd.Address(),
		RawQuery: query.Encode(),
	}

	return u.String()
}
