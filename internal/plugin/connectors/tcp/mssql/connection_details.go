package mssql

import (
	"net/url"
	"strconv"
)

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the SingleUseConnector credentials config
type ConnectionDetails struct {
	Host      string
	Port      uint
	Username  string
	Password  string
	SSLParams map[string]string
}

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

const defaultSSLMode = sslModeRequire

const defaultMSSQLPort = uint(1433)

// NewConnectionDetails is a constructor of ConnectionDetails structure from a
// map of credentials.
func NewConnectionDetails(credentials map[string][]byte) *ConnectionDetails {

	connDetails := &ConnectionDetails{}

	if len(credentials["host"]) > 0 {
		connDetails.Host = string(credentials["host"])
	}

	connDetails.Port = defaultMSSQLPort
	if len(credentials["port"]) > 0 {
		port64, _ := strconv.ParseUint(string(credentials["port"]), 10, 64)
		connDetails.Port = uint(port64)
	}

	if len(credentials["username"]) > 0 {
		connDetails.Username = string(credentials["username"])
	}

	if len(credentials["password"]) > 0 {
		connDetails.Password = string(credentials["password"])
	}

	connDetails.SSLParams = newSSLParams(credentials)

	return connDetails
}

func newSSLParams(credentials map[string][]byte) map[string]string {
	sslMode := string(credentials["sslmode"])
	params, ok := sslModeToBaseParams[sslMode]

	if !ok {
		credentials["sslmode"] = []byte(defaultSSLMode)
		return newSSLParams(credentials)
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

	return params
}

// Address returns a string representing the network location (host and port)
// of a MSSQL server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
func (cd *ConnectionDetails) address() string {
	return cd.Host + ":" + strconv.FormatUint(uint64(cd.Port), 10)
}

// URL returns a string URL from connection details
func (cd *ConnectionDetails) URL() string {
	query := url.Values{}

	for key, value := range cd.SSLParams {
		query.Add(key, value)
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cd.Username, cd.Password),
		Host:     cd.address(),
		RawQuery: query.Encode(),
	}

	return u.String()
}
