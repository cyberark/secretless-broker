package tests

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
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

	// generate TestSuiteLiveConfigurations
	RunTestCase = test.NewRunTestCase(pkg.RunQuery)

}
