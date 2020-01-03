package testutil

import (
	"log"
	"os"
	"strconv"
)

// Verbose reads the VERBOSE environment variable to determine output mode.
var Verbose = func() bool {
	debug := os.Getenv("VERBOSE")
	for _, truthyVal := range []string{"true", "yes", "t", "y"} {
		if truthyVal == debug {
			return true
		}
	}
	return false
}()

// DBConfig holds configuration information for a database
type DBConfig struct {
	HostWithTLS    string
	HostWithoutTLS string
	Port           string
	User           string
	Password       string
	Protocol       string
}

// NewDbConfigFromEnv creates a new DbConfig from ENV variables.  The variables
// assumed to exist are:
//
//     DB_HOST_TLS
//     DB_HOST_NO_TLS
//     DB_PORT
//     DB_USER
//     DB_PASSWORD
//     DB_PROTOCOL
//
func NewDbConfigFromEnv() DBConfig {

	requiredEnvVars := []string{
		"DB_HOST_TLS",
		"DB_HOST_NO_TLS",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_PROTOCOL",
	}

	// Validate they exist
	for _, field := range requiredEnvVars {
		if _, found := os.LookupEnv(field); !found {
			log.Panicf("ERROR: $%v envvar wasn't found\n", field)
		}
	}

	// Read them into the DBConfig
	return DBConfig{
		HostWithTLS:    os.Getenv("DB_HOST_TLS"),
		HostWithoutTLS: os.Getenv("DB_HOST_NO_TLS"),
		Port:           os.Getenv("DB_PORT"),
		User:           os.Getenv("DB_USER"),
		Password:       os.Getenv("DB_PASSWORD"),
		Protocol:       os.Getenv("DB_PROTOCOL"),
	}
}

var sampleDbConfig = NewDbConfigFromEnv()

// SecretlessHost gets its value from the SECRETLESS_HOST ENV variable, and
// allows us to specify a different host when doing development, for faster code
// reloading.  See the "dev" script in this folder.
var SecretlessHost = func() string {
	if host, ok := os.LookupEnv("SECRETLESS_HOST"); ok {
		return host
	}
	return "secretless"
}()

// SecretlessPort gets its value from the SECRETLESS_PORT ENV variable, and
// allows us to specify a different port to use for the secretlessy-proxy
// host.
var SecretlessPort = func() int {
	port := 2223
	if portEnv, ok := os.LookupEnv("SECRETLESS_PORT"); ok {
		var err error
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			log.Panicf("ERROR: Invalid SECRETLESS_PORT envvar: $%v\n", portEnv)
		}
	}
	return port
}()

func init() {
	// TEST_ROOT is used to direct where secretless.yml gets generated
	testRoot, ok := os.LookupEnv("TEST_ROOT")
	if !ok {
		panic("ERROR: $TEST_ROOT envvar wasn't found\n")
	}

	os.Chdir(testRoot)
}
