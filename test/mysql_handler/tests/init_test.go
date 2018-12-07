package tests

import (
	"fmt"
	"net"

	. "github.com/cyberark/secretless-broker/test/mysql_handler/pkg"
)

// If the SecretlessHost is unavailable, bail out...
//
func init() {
	_, err := net.LookupIP(SecretlessHost)
	if err != nil {
		fmt.Printf("ERROR: The secretless host '%s' wasn't found\n", SecretlessHost)
		panic(err)
	}

	// generate TestSuiteLiveConfigurations
	_, TestSuiteLiveConfigurations = GenerateConfigurations()
}
