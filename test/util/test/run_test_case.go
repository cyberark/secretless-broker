package test

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
)


// TODO: make ConnectionParams an interface and pass it to RunTestCase
// all flag generation and combination can happen here
func NewRunTestCase(runQuery RunQueryType) RunTestCaseType {
	_, testSuiteConfigurations := GenerateConfigurations()

	return func (testCase TestCase) {
		var expectation = "throws"
		if testCase.ShouldPass {
			expectation = "succeeds"
		}

		convey.Convey(fmt.Sprintf("%s: %s", expectation, testCase.Description), func() {
			// TODO: possibly move this logic into testCase
			liveConfiguration := testSuiteConfigurations.Find(testCase.AbstractConfiguration)

			cmdOut, err := runQuery(testCase.ClientConfiguration, liveConfiguration.ConnectionPort)

			if testCase.ShouldPass {
				convey.So(err, convey.ShouldBeNil)
			} else {
				convey.So(err, convey.ShouldNotBeNil)
			}

			if testCase.CmdOutput != nil {
				convey.So(cmdOut, convey.ShouldContainSubstring, *testCase.CmdOutput)
			}

		})
	}
}

// Flags is an array of strings passed directly to the database CLI. Eg:
//
//     []string{"-u test", "--password=wrongpassword"}
//
// allows us to treat queries in mysql and postgres via a common abstraction
type RunQueryType func(ClientConfiguration, ConnectionPort) (string, error)
type RunTestCaseType func(testCase TestCase)
