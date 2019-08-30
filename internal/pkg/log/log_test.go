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

type logMethod func(...interface{})
type logMethodF func(string, ...interface{})

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

func (lt *LogTest) LogMethodsF() map[string]logMethodF {
	return map[string]logMethodF{
		"Debugf": lt.logger.Debugf,
		"Infof": lt.logger.Infof,
		"Warnf": lt.logger.Warnf,
		"Errorf": lt.logger.Errorf,
		"Panicf": lt.logger.Panicf,
	}
}

func (lt *LogTest) LogMethods() map[string]logMethod {
	return map[string]logMethod{
		"Debug": lt.logger.Debug,
		"Debugln": lt.logger.Debugln,
		"Info": lt.logger.Info,
		"Infoln": lt.logger.Infoln,
		"Warn": lt.logger.Warn,
		"Warnln": lt.logger.Warnln,
		"Error": lt.logger.Error,
		"Errorln": lt.logger.Errorln,
		"Panic": lt.logger.Panic,
		"Panicln": lt.logger.Panicln,
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
	for methName, meth := range lt.LogMethodsF() {
		lt.ResetBuffer()
		t.Run(
			lt.testName(methName),
			func(t *testing.T) {
				meth(lt.formatStr(), lt.args()...)
				assert.Regexp(t, lt.expectedOutput(methName), lt.CurrentOutput())
			},
		)
	}

	// Unformatted methods
	for methName, meth := range lt.LogMethods() {
		lt.ResetBuffer()
		t.Run(
			lt.testName(methName),
			func(t *testing.T) {
				meth(lt.args()...)
				assert.Regexp(t, lt.expectedOutput(methName), lt.CurrentOutput())
			},
		)
	}
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
