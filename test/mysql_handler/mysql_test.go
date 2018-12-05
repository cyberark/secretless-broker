package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const SecretlessHost = "secretless"

func init() {
	_, err := net.LookupIP(SecretlessHost)
	if err != nil {
		panic(err)
	}
}

func runTestQuery(flags []string) (string, error) {
	args := []string{ "-e", "select count(*) from testdb.test" }

	// non-separated flags: ["-u test", "--ssl-mode=DISABLE"]
	for _, v := range flags {
		args = append(args, v)
	}

	log.Println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmdOut, err := exec.Command("mysql", args...).CombinedOutput()

	if err != nil {
		fmt.Println("ERROR: ", err.Error())
		fmt.Println("OUTPUT: ", string(cmdOut))
	}

	return string(cmdOut), err
}

func TestUnixSocketMySQLHandler(t *testing.T) {
	Convey("Connect over a UNIX socket", t, func() {
		testCases := SharedTestCases()

		for testName, testCase := range testCases {
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
		testCases := SharedTestCases()

		for testName, testCase := range testCases {
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

func stringPointer(s string) *string {
	return &s
}
// TestCase represents the conditions and expected outcomes of a test
//
// For AssertFailure, we assume success without explicit failure
//
// For CmdOutput, there are two cases we need:
// 1. Don't assert on the command output
// 2. Assert the command output is empty, or otherwise
// A string pointer distinguishes between those cases
type TestCase struct {
	Flags         []string
	AssertFailure bool
	CmdOutput     *string
}
func SharedTestCases() map[string]TestCase  {
	testCases := map[string]TestCase{
		"With username, wrong password": {
			Flags: []string{
				"--user=testuser",
				"--password=wrongpassword",
			},
			CmdOutput: stringPointer("2"),
		},
		"With wrong username, wrong password": {
			Flags: []string{
				"--user=wrongusername",
				"--password=wrongpassword",
			},
			CmdOutput: stringPointer("2"),
		},
		"With empty username, empty password": {
			Flags: []string{
				"--user=",
				"--password=",
			},
			CmdOutput: stringPointer("2"),
		},
		"Client is not able to connect to Secretless via TLS": {
			Flags: []string{
				"--user=",
				"--password=",
				"--ssl",
			},
			AssertFailure: true,
		},
	}

	return testCases
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
