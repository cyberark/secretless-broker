package pkg

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/smartystreets/goconvey/convey"
)

func Runner(testCase TestCase) {
	var expectation = "succeeds"
	if testCase.AssertFailure {
		expectation = "throws"
	}

	convey.Convey(fmt.Sprintf("%s: %s", expectation, testCase.Description), func() {
		liveConfiguration := TestSuiteLiveConfigurations.Find(testCase.AbstractConfiguration)

		testCase.Flags = append(testCase.Flags, liveConfiguration.ConnectionFlags()...)

		cmdOut, err := runQuery(testCase.Flags)

		if testCase.AssertFailure {
			convey.So(err, convey.ShouldNotBeNil)
		} else {
			convey.So(err, convey.ShouldBeNil)
			convey.So(cmdOut, convey.ShouldContainSubstring, successOutput)
		}

		if testCase.CmdOutput != nil {
			convey.So(cmdOut, convey.ShouldContainSubstring, *testCase.CmdOutput)
		}

	})
}


// Flags is an array of strings passed directly to the mysql CLI. Eg:
//
//     []string{"-u test", "--password=wrongpassword"}
//
func runQuery(flags []string) (string, error) {
	args := []string{"-e", "select count(*) from testdb.test"}

	for _, v := range flags {
		args = append(args, v)
	}

	// Pre command logs
	convey.Println("")
	convey.Println("---<< EXECUTED")
	convey.Println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmdOut, err := exec.Command("mysql", args...).CombinedOutput()

	// Post command logs
	if Verbose {
		if err != nil {
			convey.Println("--->> RESULTS")
			convey.Println("----- ERROR: ")
			convey.Println(err.Error())
		}
		convey.Println("----- OUTPUT: ")
		convey.Println(string(cmdOut))
	}
	convey.Println("---<< END")
	convey.Println("")

	return string(cmdOut), err
}

