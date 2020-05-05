package tests

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mysql/pkg"
	"github.com/cyberark/secretless-broker/test/util/testutil"
)

var RunTestCase testutil.RunTestCaseType

// If the SecretlessHost is unavailable, bail out...
//
func init() {
	_, err := net.LookupIP(testutil.SecretlessHost)
	if err != nil {
		fmt.Printf("ERROR: The secretless host '%s' wasn't found\n", testutil.SecretlessHost)
		panic(err)
	}

	// generate TestSuiteLiveConfigurations
	RunTestCase = testutil.NewRunTestCase(pkg.RunQuery)

}
