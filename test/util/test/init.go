package test

import (
	"fmt"
	"os"
)

// ENV Configuration: Verbose output mode
//
var Verbose = func() bool {
	debug := os.Getenv("VERBOSE")
	for _, truthyVal := range []string{"true", "yes", "t", "y"} {
		if truthyVal == debug {
			return true
		}
	}
	return false
}()

// ENV Configuration: DBConfig
//
type TestDBConfigType struct {
	DB_HOST_TLS string
	DB_HOST_NO_TLS string
	DB_PORT string
	DB_USER string
	DB_PASSWORD string
	DB_PROTOCOL string
}
var TestDBConfig = func() TestDBConfigType {
	fields := []string{
		"DB_HOST_TLS",
		"DB_HOST_NO_TLS",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_PROTOCOL",
	}

	for _, field := range fields {
		if _, found := os.LookupEnv(field); !found  {
			fmt.Printf("ERROR: $%v envvar wasn't found\n", field)
			panic("$" + field)
		}
	}

	return TestDBConfigType{
		DB_HOST_TLS: os.Getenv("DB_HOST_TLS"),
		DB_HOST_NO_TLS: os.Getenv("DB_HOST_NO_TLS"),
		DB_PORT: os.Getenv("DB_PORT"),
		DB_USER: os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_PROTOCOL: os.Getenv("DB_PROTOCOL"),
	}
}()

// ENV Configuration: Name of Secretless host to use
//
// Allows us to specify a different host when doing development, for
// faster code reloading.  See the "dev" script in this folder.
//
var SecretlessHost = func() string {
	if host, ok := os.LookupEnv("SECRETLESS_HOST"); ok {
		return host
	}
	return "secretless"
}()

func init() {
	// TEST_ROOT is used to direct where secretless.yml gets generated
	testRoot, ok := os.LookupEnv("TEST_ROOT")
	if !ok {
		fmt.Printf("ERROR: $TEST_ROOT envvar wasn't found\n")
		panic("$TEST_ROOT")
	}

	os.Chdir(testRoot)
}
