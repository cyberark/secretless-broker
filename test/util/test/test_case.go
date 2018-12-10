package test

// TestDefinition allows us to treat similar tests with variations as data.
//
// By default, a TestDefinition is assumed not to error.  When we expect
// an error, however, we can set ShouldPass = true.
//
// For CmdOutput, there are two cases we need:
//
// 1. Make no assertions on the command output (nil)
// 2. Assert the command output is empty, or some specific string.
//
// A string pointer, with its possible nil value, lets us distinguish
// those cases. A string would not.
//
type TestDefinition struct {
	Description string
	ClientConfiguration
	ShouldPass  bool
	CmdOutput   *string
}

type TestCase struct {
	TestDefinition
	AbstractConfiguration
}
