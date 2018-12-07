package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

var TestDir string
func init() {
	secretlessRoot, ok := os.LookupEnv("PROJECT_ROOT")
	if !ok {
		fmt.Printf("ERROR: Secretless $PROJECT_ROOT envvar wasn't found\n")
		panic("$PROJECT_ROOT")
	}

	TestDir = filepath.Join(secretlessRoot, "./test/mysql_handler")
	os.Chdir(TestDir)

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
