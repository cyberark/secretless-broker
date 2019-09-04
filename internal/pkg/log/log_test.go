package log

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	logapi "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/stretchr/testify/assert"
)

func TestDebugEnabled(t *testing.T) {
	assert.True(t, New(true).DebugEnabled())
	assert.True(t, NewForService("abc", true).DebugEnabled())
	assert.True(t, NewWithOptions(&bytes.Buffer{}, "abc", true).DebugEnabled())

	assert.False(t, New(false).DebugEnabled())
	assert.False(t, NewForService("abc", false).DebugEnabled())
	assert.False(t, NewWithOptions(&bytes.Buffer{}, "abc", false).DebugEnabled())
}

// TestAllOutputMethods tests the logger's output-generating methods over the
// logger's 4 possible states.
func TestAllOutputMethods(t *testing.T) {
	test := NewLogTest(true, "prefix")
	test.RunAllTests(t)

	test = NewLogTest(true, "")
	test.RunAllTests(t)

	test = NewLogTest(false, "prefix")
	test.RunAllTests(t)

	test = NewLogTest(false, "")
	test.RunAllTests(t)
}

// Methods and types for describing and classifying the methods of a Logger

// Logger methods have two possible signatures: one for "formatted" methods
// that end in "f" and use "Printf" style format strings, and for normal methods
// that just print their arguments
type logMethod func(...interface{})
type logMethodF func(string, ...interface{})

// methodName elevates the string name of log methods to a proper type. This
// is good because the behavior of the various log methods can be inferred from
// their names.
type methodName string

// Names ending in "f" require a format string
func (name methodName) requiresFormatString()  bool {
	// Only formatted methods end in the letter f
	formattedRe := regexp.MustCompile("f$")
	return formattedRe.MatchString(string(name))
}

// debugOnly identifies methods that produce output only when the Logger is in
// debug mode.
func (name methodName) debugOnly()  bool {
	return strings.HasPrefix(string(name), "Debug") ||
		strings.HasPrefix(string(name), "Info")
}

// LogTest represents a full test of all output-generating methods on a Logger.
type LogTest struct {
	logger logapi.Logger
	backingBuffer *bytes.Buffer
}

func NewLogTest(isDebug bool, prefix string) *LogTest {
	backingBuffer := &bytes.Buffer{}
	logger := NewWithOptions(backingBuffer, prefix, isDebug)

	return &LogTest{
		logger: logger,
		backingBuffer: backingBuffer,
	}
}

func (lt *LogTest) RunAllTests(t *testing.T) {

	// Format strings and sample arguments used in the test cases
	const testCaseFormatStr = "aaa %s bbb %d ccc %2.1f ddd \t eee"
	testCaseArgs := []interface{}{ "stringval", 123, 1.234 }

	// Formatted methods
	for methodName, method := range map[string]logMethodF{
		"Debugf": lt.logger.Debugf,
		"Infof": lt.logger.Infof,
		"Warnf": lt.logger.Warnf,
		"Errorf": lt.logger.Errorf,
		"Panicf": lt.logger.Panicf,
	} {
		lt.ResetBuffer()
		t.Run(
			lt.descriptionForTest(methodName),
			func(t *testing.T) {
				method(testCaseFormatStr, testCaseArgs...)
				assert.Regexp(t, lt.expectedOutput(methodName), lt.CurrentOutput())
			},
		)
	}

	// Unformatted methods
	for methodName, method := range map[string]logMethod{
		"Debug":   lt.logger.Debug,
		"Debugln": lt.logger.Debugln,
		"Info":    lt.logger.Info,
		"Infoln":  lt.logger.Infoln,
		"Warn":    lt.logger.Warn,
		"Warnln":  lt.logger.Warnln,
		"Error":   lt.logger.Error,
		"Errorln": lt.logger.Errorln,
		"Panic":   lt.logger.Panic,
		"Panicln": lt.logger.Panicln,
	} {
		lt.ResetBuffer()
		t.Run(
			lt.descriptionForTest(methodName),
			func(t *testing.T) {
				method(testCaseArgs...)
				assert.Regexp(t, lt.expectedOutput(methodName), lt.CurrentOutput())
			},
		)
	}
}

func (lt *LogTest) expectedOutput(methNameStr string) *regexp.Regexp {
	methodName := methodName(methNameStr)

	// Debug methods produce no output unless debug is enabled
	if methodName.debugOnly() && !lt.logger.DebugEnabled() {
		return regexp.MustCompile("")
	}

	datetimeRe := `\d{4}\/\d{2}\/\d{2} \d{1,2}:\d{2}:\d{2}`

	// expected content is different for formatted and unformatted methods
	methodResultRe := `stringval 123 1\.234`
	if methodName.requiresFormatString() {
		methodResultRe = `aaa stringval bbb 123 ccc 1\.2 ddd\s{1,8}eee`
	}

	// expected content also changes if there is a prefix
	fullLineRe := "^"+datetimeRe+" "+methodResultRe+"\n"
	if lt.logger.Prefix() != "" {
		prefix := lt.logger.Prefix()
		fullLineRe = "^"+datetimeRe+" "+prefix+": "+methodResultRe+"\n"
	}

	return regexp.MustCompile(fullLineRe)
}

func (lt *LogTest) descriptionForTest(methodName string) string {
	return fmt.Sprintf(
		"%s/prefix='%s'/isDebug=%t",
		methodName,
		lt.logger.Prefix(),
		lt.logger.DebugEnabled(),
	)
}

func (lt *LogTest) ResetBuffer() {
	lt.backingBuffer.Reset()
}

func (lt *LogTest) CurrentOutput() string {
	return lt.backingBuffer.String()
}
