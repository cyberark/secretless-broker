package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TODO: move this to a util folder somewhere
func stringPointer(s string) *string {
	return &s
}

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

// If the SecretlessHost is unavailable, bail out...
//
func init() {
	_, err := net.LookupIP(SecretlessHost)
	if err != nil {
		fmt.Printf("ERROR: The secretless host '%s' wasn't found\n", SecretlessHost)
		panic(err)
	}
}

// TestCase allows us to treat similar tests with variations as data.
//
// By default, a TestCase is assumed not to error.  When we expect
// an error, however, we can set AssertFailure = true.
//
// For CmdOutput, there are two cases we need:
//
// 1. Make no assertions on the command output (nil)
// 2. Assert the command output is empty, or some specific string.
//
// A string pointer, with its possible nil value, lets us distinguish
// those cases. A string would not.
//
type TestCase struct {
	Flags         []string
	AssertFailure bool
	CmdOutput     *string
}
var SharedTestCases = func() map[string]TestCase  {
	genericOutput := stringPointer("2")
	return map[string]TestCase{
		"With username, wrong password": {
			Flags: []string{
				"--user=testuser",
				"--password=wrongpassword",
			},
			CmdOutput: genericOutput,
		},
		"With wrong username, wrong password": {
			Flags: []string{
				"--user=wrongusername",
				"--password=wrongpassword",
			},
			CmdOutput: genericOutput,
		},
		"With empty username, empty password": {
			Flags: []string{
				"--user=",
				"--password=",
			},
			CmdOutput: genericOutput,
		},
		"Client is not able to connect to Secretless via TLS": {
			Flags: []string{
				"--user=",
				"--password=",
				"--ssl-verify-server-cert=TRUE",
				"--ssl",
			},
			AssertFailure: true,
			CmdOutput: stringPointer("ERROR 2026 (HY000): SSL connection error: SSL is required, but the server does not support"),
		},
	}
}()

// Flags is an array of strings passed directly to the mysql CLI. Eg:
//
//     []string{"-u test", "--password=wrongpassword"}
//
func runTestQuery(flags []string) (string, error) {
	args := []string{"-e", "select count(*) from testdb.test"}

	for _, v := range flags {
		args = append(args, v)
	}

	// Pre command logs
	Println("")
	Println("---<< EXECUTED")
	Println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmdOut, err := exec.Command("mysql", args...).CombinedOutput()

	// Post command logs
	if Verbose {
		if err != nil {
			Println("--->> RESULTS")
			Println("----- ERROR: ")
			Println(err.Error())
		}
		Println("----- OUTPUT: ")
		Println(string(cmdOut))
	}
	Println("---<< END")
	Println("")

	return string(cmdOut), err
}

func TestUnixSocketMySQLHandler(t *testing.T) {
	Convey("Connect over a UNIX socket", t, func() {

		for testName, testCase := range SharedTestCases {
			Convey(testName, func() {

				testCase.Flags = append(testCase.Flags, "--socket=sock/mysql.sock")
				cmdOut, err := runTestQuery(testCase.Flags)

				if testCase.AssertFailure {
					So(err, ShouldNotBeNil)
				} else {
					So(err, ShouldBeNil)
				}

				if testCase.CmdOutput != nil {
					So(cmdOut, ShouldContainSubstring, *testCase.CmdOutput)
				}
			})
		}
	})
}

func TestTCPMySQLHandler(t *testing.T) {
	Convey("Connect over TCP secretless->server TLS support and sslmode default", t, func() {

		for testName, testCase := range SharedTestCases{
			Convey(testName, func() {
				testCase.Flags = append(testCase.Flags, "--port=3306")
				testCase.Flags = append(testCase.Flags, fmt.Sprintf("--host=%s", SecretlessHost))
				cmdOut, err := runTestQuery(testCase.Flags)

				if testCase.AssertFailure {
					So(err, ShouldNotBeNil)
				} else {
					So(err, ShouldBeNil)
				}

				if testCase.CmdOutput != nil {
					So(cmdOut, ShouldContainSubstring, *testCase.CmdOutput)
				}

			})
		}
	})
}



func TestTLSMySQLHandler(t *testing.T) {
	Convey("TLS: sslmode default", t, func() {

		Convey("sslmode default", func() {
			Convey("Connect to server with TLS support", func() {
				cmdOut, err := runTestQuery(
					[]string{
						"--user=",
						"--password=",
						fmt.Sprintf("--host=%s", SecretlessHost),
						"--port=3306",
					},
				)

				So(err, ShouldBeNil)
				So(cmdOut, ShouldContainSubstring, "2")
			})

			Convey("Fail to connect to server without TLS support", func() {
				cmdOut, err := runTestQuery(
					[]string{
						"--user=",
						"--password=",
						fmt.Sprintf("--host=%s", SecretlessHost),
						"--port=4306",
					},
				)

				So(err, ShouldNotBeNil)
				So(cmdOut, ShouldContainSubstring, "ERROR 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it")
			})
		})
	})
}
