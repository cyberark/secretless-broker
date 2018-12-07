package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

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
