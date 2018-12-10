package tests

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/test/pg2_handler/pkg"
	"github.com/cyberark/secretless-broker/test/util/test"
)


var RunTestCase test.RunTestCaseType
// If the SecretlessHost is unavailable, bail out...
//
func init() {
	_, err := net.LookupIP(test.SecretlessHost)
	if err != nil {
		fmt.Printf("ERROR: The secretless host '%s' wasn't found\n", test.SecretlessHost)
		panic(err)
	}

	RunTestCase = test.NewRunTestCase(pkg.RunQuery)

}
