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

var allMethods = []string{
	"Debug", "Debugf", "Debugln",
	"Info", "Infof", "Infoln",
	"Warn", "Warnf", "Warnln",
	"Error", "Errorf", "Errorln",
	"Panic", "Panicf", "Panicln",
}

var formattedMethods = []string{
	"Debugf", "Infof", "Warnf", "Errorf", "Panicf",
}

func isFormattedMethod(methName string) bool {
	ret := false
	for _, meth := range formattedMethods {
		if methName == meth {
			ret = true
		}
	}
	return ret
}

func isDebugOnlyMethod(methName string) bool {
	return strings.HasPrefix(methName, "Debug") ||
		strings.HasPrefix(methName, "Info")
}

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

func (lt *LogTest) ResetBuffer() {
	lt.backingBuffer.Reset()
}

func (lt *LogTest) CurrentOutput() string {
	return lt.backingBuffer.String()
}


func (lt *LogTest) RunAllTests(t *testing.T) {

	// Formatted methods
	lt.ResetBuffer()
	t.Run(
		lt.testName("Debugf"),
		func(t *testing.T) {
			lt.logger.Debugf(lt.formatStr(), lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Debugf"), lt.CurrentOutput())
		},
	)

	lt.ResetBuffer()
	t.Run(
		lt.testName("Infof"),
		func(t *testing.T) {
			lt.logger.Infof(lt.formatStr(), lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Infof"), lt.CurrentOutput())
		},
	)

	lt.ResetBuffer()
	t.Run(
		lt.testName("Warnf"),
		func(t *testing.T) {
			lt.logger.Warnf(lt.formatStr(), lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Warnf"), lt.CurrentOutput())
		},
	)

	lt.ResetBuffer()
	t.Run(
		lt.testName("Errorf"),
		func(t *testing.T) {
			lt.logger.Errorf(lt.formatStr(), lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Errorf"), lt.CurrentOutput())
		},
	)

	lt.ResetBuffer()
	t.Run(
		lt.testName("Panicf"),
		func(t *testing.T) {
			lt.logger.Panicf(lt.formatStr(), lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Panicf"), lt.CurrentOutput())
		},
	)

	// Unformatted methods
	lt.ResetBuffer()
	t.Run(
		lt.testName("Debugln"),
		func(t *testing.T) {
			lt.logger.Debugln(lt.args()...)
			assert.Regexp(t, lt.expectedOutput("Debugln"), lt.CurrentOutput())
		},
	)
}

func (lt *LogTest) formatStr() string {
	return "aaa %s bbb %d ccc %2.1f ddd \t eee"
}

// testArgsF returns sample arguments for formatting methods ending with an "f"
func (lt *LogTest) args() []interface{} {
	return []interface{}{ "stringval", 123, 1.234 }
}

func (lt *LogTest) expectedOutput(methName string) *regexp.Regexp {
	// Debug methods produce no output unless debug is enabled
	if isDebugOnlyMethod(methName) && !lt.logger.DebugEnabled() {
		return regexp.MustCompile("")
	}

	datetimeRe := `\d{4}\/\d{2}\/\d{2} \d{1,2}:\d{2}:\d{2}`

	// expected content is different for formatted and unformatted methods
	methodResultRe := `stringval 123 1\.234`
	if isFormattedMethod(methName) {
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

func (lt *LogTest) testName(methName string) string {
	return fmt.Sprintf(
		"%s/prefix='%s'/isDebug=%t",
		methName,
		lt.logger.Prefix(),
		lt.logger.DebugEnabled(),
	)
}
