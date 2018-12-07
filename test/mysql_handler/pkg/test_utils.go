package pkg

// TestData allows us to treat similar tests with variations as data.
//
// By default, a TestData is assumed not to error.  When we expect
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
type TestData struct {
	Description string
	Flags         []string
	AssertFailure bool
	CmdOutput     *string
}

type TestCase struct {
	AbstractConfiguration
	TestData
}

// set by importer of this file
var TestSuiteLiveConfigurations LiveConfigurations
const successOutput = "2"
