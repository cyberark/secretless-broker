package tests

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
	"github.com/cyberark/secretless-broker/test/util/test"
)


var MyRunTestCase test.RunTestCase
// If the SecretlessHost is unavailable, bail out...
//
func init() {
	_, err := net.LookupIP(test.SecretlessHost)
	if err != nil {
		fmt.Printf("ERROR: The secretless host '%s' wasn't found\n", test.SecretlessHost)
		panic(err)
	}

	// generate TestSuiteLiveConfigurations
	_, testSuiteLiveConfigurations := test.GenerateConfigurations()
	MyRunTestCase = test.NewRunTestCase(pkg.RunQuery, testSuiteLiveConfigurations)

}
