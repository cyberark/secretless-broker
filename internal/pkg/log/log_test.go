package log

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var formattedLoggingMethods = []string{
	"Debugf",
	"Infof",
	"Warnf",
	"Errorf",
	"Panicf",
}

var unformattedLoggingMethods = []string{
	"Debug",
	"Debugln",
	"Info",
	"Infoln",
	"Warn",
	"Warnln",
	"Error",
	"Errorln",
	"Panic",
	"Panicln",
}

type OutputTest struct {
	name         string
	outputMethod reflect.Value
	outputBuffer *bytes.Buffer
}

// Datetime string should look something like: `2019/01/01 13:14:15`
var datetimeRegexString = `\d{4}\/\d{2}\/\d{2} \d{1,2}:\d{2}:\d{2}`

// Formatted string should include a string, int, and a float working
var formattedArgsResultsRegexString = `aaa stringval bbb 123 ccc 1\.2 ddd\s{1,8}eee`

// Unformatted string should include a string, int, and a float working
var unformattedArgsResultsRegexString = `stringval 123 1\.234 aaa`

type ExpectedEntry struct {
	prefix          string
	expectedContent *regexp.Regexp
}

func NewExpectedEntry(prefix string, contentRegexString string) *ExpectedEntry {
	return &ExpectedEntry{
		prefix:          prefix,
		expectedContent: regexp.MustCompile(contentRegexString),
	}
}

var formattedPrefixMatchers = []*ExpectedEntry{
	NewExpectedEntry("",
		"^"+datetimeRegexString+" "+formattedArgsResultsRegexString+"\n"),
	NewExpectedEntry("prefix",
		"^"+datetimeRegexString+" prefix: "+formattedArgsResultsRegexString+"\n"),
}

var unformattedPrefixMatchers = []*ExpectedEntry{
	NewExpectedEntry("",
		"^"+datetimeRegexString+" "+unformattedArgsResultsRegexString+"\n"),
	NewExpectedEntry("prefix",
		"^"+datetimeRegexString+" prefix: "+unformattedArgsResultsRegexString+"\n"),
}

func newOutputTest(methodName string, isDebug bool, prefix string) OutputTest {
	outputBuffer := &bytes.Buffer{}
	logger := NewWithOptions(outputBuffer, prefix, isDebug)

	loggerType := reflect.ValueOf(logger)
	methodPointer := loggerType.MethodByName(methodName)

	return OutputTest{
		name:         methodName,
		outputMethod: methodPointer,
		outputBuffer: outputBuffer,
	}
}

func TestDebugEnabled(t *testing.T) {
	assert.True(t, New(true).DebugEnabled())
	assert.True(t, NewForService("abc", true).DebugEnabled())
	assert.True(t, NewWithOptions(&bytes.Buffer{}, "abc", true).DebugEnabled())

	assert.False(t, New(false).DebugEnabled())
	assert.False(t, NewForService("abc", false).DebugEnabled())
	assert.False(t, NewWithOptions(&bytes.Buffer{}, "abc", false).DebugEnabled())
}

func TestFormattedLogging(t *testing.T) {
	// Iterate over prefixes and their string representation of the
	// corresponding regexes
	for _, prefixMatcher := range formattedPrefixMatchers {
		// Iterate over all the logging methods that accept formatting string
		// as an argument
		for _, methodName := range formattedLoggingMethods {

			// Iterate over both true and false values of the debug flag
			for _, isDebug := range []bool{true, false} {
				// Use the specified prefix, flag, and method name to create a struct
				// containing all the information needed to run the specific test for
				// a combination of those parameters
				testCase := newOutputTest(methodName, isDebug, prefixMatcher.prefix)

				// Create a unique test case for this iteration
				t.Run(fmt.Sprintf("%s/prefix='%s'/isDebug=%t", testCase.name,
					prefixMatcher.prefix, isDebug), func(t *testing.T) {

					// Create a list of arguments that we will send to the formatted
					// logging method. We must use reflect.Value since that's what
					// the method pointer will expect
					args := []reflect.Value{
						reflect.ValueOf("aaa %s bbb %d ccc %2.1f ddd \t eee"),
						reflect.ValueOf("stringval"),
						reflect.ValueOf(123),
						reflect.ValueOf(1.234),
					}

					// Invoke the specified `logger.<methodName>` with the arguments
					// defined above
					testCase.outputMethod.Call(args)

					if !isDebug &&
						(strings.HasPrefix(methodName, "Debug") ||
							strings.HasPrefix(methodName, "Info")) {
						// If we have the debug flag on, expect that debug and
						// info messages don't show up
						assert.Empty(t, testCase.outputBuffer.String())
					} else {
						// Otherwise assert that the printed output matches
						// the defined regex
						assert.Regexp(t, prefixMatcher.expectedContent, testCase.outputBuffer.String())
					}
				})
			}
		}
	}
}

func TestUnformattedLogging(t *testing.T) {
	// Iterate over prefixes and their string representation of the
	// corresponding regexes
	for _, prefixMatcher := range unformattedPrefixMatchers {
		// Iterate over all the logging methods that accept unformatted list
		// of arguments
		for _, methodName := range unformattedLoggingMethods {

			// Iterate over both true and false values of the debug flag
			for _, isDebug := range []bool{true, false} {
				// Use the specified prefix, flag, and method name to create a struct
				// containing all the information needed to run the specific test for
				// a combination of those parameters
				testCase := newOutputTest(methodName, isDebug, prefixMatcher.prefix)

				// Create a unique test case for this iteration
				t.Run(fmt.Sprintf("%s/prefix='%s'/isDebug=%t", testCase.name,
					prefixMatcher.prefix, isDebug), func(t *testing.T) {

					// Create a list of arguments that we will send to the unformatted
					// logging method. We must use reflect.Value since that's what
					// the method pointer will expect
					args := []reflect.Value{
						reflect.ValueOf("stringval"),
						reflect.ValueOf(123),
						reflect.ValueOf(1.234),
						reflect.ValueOf("aaa"),
					}

					// Invoke the specified `logger.<methodName>` with the arguments
					// defined above
					testCase.outputMethod.Call(args)

					if !isDebug &&
						(strings.HasPrefix(methodName, "Debug") ||
							strings.HasPrefix(methodName, "Info")) {
						// If we have the debug flag on, expect that debug and
						// info messages don't show up
						assert.Empty(t, testCase.outputBuffer.String())
					} else {
						// Otherwise assert that the printed output matches
						// the defined regex
						assert.Regexp(t, prefixMatcher.expectedContent, testCase.outputBuffer.String())
					}
				})
			}
		}
	}
}
