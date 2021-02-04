package testutil

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
)

// NewRunTestCase returns a function that, given a TestCase, will run the unit
// tests specified by that TestCase, using the query type given by RunQueryType.
// TODO: make ConnectionParams an interface and pass it to RunTestCase
func NewRunTestCase(runQuery RunQueryType) RunTestCaseType {
	_, testSuiteConfigurations := GenerateConfigurations()

	return func(testCase TestCase) {
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

// FindLiveConfiguration takes AbstractConfiguration and finds a live configuration for
// an active Secretless endpoint
func FindLiveConfiguration(ac AbstractConfiguration) LiveConfiguration {
	_, testSuiteConfigurations := GenerateConfigurations()

	return testSuiteConfigurations.Find(ac)
}

// RunQueryType represents a function that takes in database credentials
// and options, uses them to execute a test query, and returns the output
// of that query.
type RunQueryType func(ClientConfiguration, ConnectionPort) (string, error)

// RunTestCaseType represents a function for executing the unit tests
// specified by a TestCase.
type RunTestCaseType func(testCase TestCase)
