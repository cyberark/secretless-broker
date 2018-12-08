package test

import (
	"fmt"
	"io/ioutil"
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

// ENV Configuration: Database protocol
//
var DBProtocol = func() string {
	dBProtocol := os.Getenv("TEST_DB_PROTOCOL")
	if dBProtocol == "" {
		fmt.Printf("ERROR: $TEST_DB_PROTOCOL envvar wasn't found\n")
		panic("$TEST_DB_PROTOCOL")
	}

	return dBProtocol
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

// TODO: explain the reasoning behind the below
// NOTE: fixtures are generated in bash script
// this requires coordination between the bash and the Go code
func init() {
	testRoot, ok := os.LookupEnv("TEST_ROOT")
	if !ok {
		fmt.Printf("ERROR: $TEST_ROOT envvar wasn't found\n")
		panic("$TEST_ROOT")
	}

	os.Chdir(testRoot)

	// set valid-ca fixture
	validCABytes, err := ioutil.ReadFile("./fixtures/valid-ca.pem")
	if err != nil {
		fmt.Printf("ERROR: valid-ca.pem wasn't found\n")
		panic(err)
	}
	Valid = SSLRootCertType(validCABytes)

	// set invalid-ca fixture
	invalidCABytes, err := ioutil.ReadFile("./fixtures_static/invalid-ca.pem")
	if err != nil {
		fmt.Printf("ERROR: invalid-ca.pem wasn't found\n")
		panic(err)
	}
	Invalid = SSLRootCertType(invalidCABytes)
}
