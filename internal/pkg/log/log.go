package log

import (
	"io"
	stdlib_log "log"
	"os"

	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
)

var defaultOutputBuffer = os.Stdout

// Logger is the main logging object that can be used to log messages to stdout
// or any other io.Writer. Delegates to `log.Logger` for writing to the buffer.
type Logger struct {
	BackingLogger *stdlib_log.Logger
	IsDebug       bool
	prefix        string
}

// severity is an integer representation of the severity level associated with
// a logging message.
type severity uint8

const (
	// DebugSeverity indicates a debug logging message
	DebugSeverity severity = iota
	// InfoSeverity indicates an informational logging message
	InfoSeverity
	// WarnSeverity indicates a warning logging message
	WarnSeverity
	// ErrorSeverity indicates a critical severity logging message
	ErrorSeverity
	// PanicSeverity indicates a severity logging message that is unliekly to be
	// recovered from
	PanicSeverity
)

// severityLevels is a list of all available severity levels. This is a shorthand
// to work around the fact that Golang cannot `range` over an enum.
var severityLevels = []severity{
	DebugSeverity,
	InfoSeverity,
	WarnSeverity,
	ErrorSeverity,
	PanicSeverity,
}

// New method instantiates a new logger that we can write things to.
func New(isDebug bool) log_api.Logger {
	return NewWithOptions(defaultOutputBuffer, "", isDebug)
}

// NewForService method instantiates a new logger that includes information about
// the service itself.
func NewForService(serviceName string, isDebug bool) log_api.Logger {
	return NewWithOptions(defaultOutputBuffer, serviceName, isDebug)
}

// NewWithOptions method instantiates a new logger with all configurable options.
// This specific constructor is not intended to be used directly by clients.
func NewWithOptions(outputBuffer io.Writer, prefix string, isDebug bool) log_api.Logger {
	return &Logger{
		BackingLogger: stdlib_log.New(outputBuffer, "", stdlib_log.LstdFlags),
		IsDebug:       isDebug,
		prefix:        prefix,
	}
}

func (logger *Logger) shouldPrint(severityLevel severity) bool {
	if !logger.IsDebug && (severityLevel == InfoSeverity || severityLevel == DebugSeverity) {
		return false
	}

	return true
}

func prependString(prependString string, args ...interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = prependString
	for idx, val := range args {
		newArgs[idx+1] = val
	}

	return newArgs
}

// DebugEnabled returns if the debug logging should be displayed for a particular
// logger instance
func (logger *Logger) DebugEnabled() bool {
	return logger.IsDebug
}

// Prefix returns the prefix that will be prepended to all output messages
func (logger *Logger) Prefix() string {
	return logger.prefix
}

// ---------------------------
// Main logging methods that funnel all the info here

func (logger *Logger) logf(severityLevel severity, format string, args ...interface{}) {
	if !logger.shouldPrint(severityLevel) {
		return
	}

	if logger.prefix != "" {
		format = "%s: " + format
		args = prependString(logger.prefix, args...)
	}

	logger.BackingLogger.Printf(format, args...)
}

func (logger *Logger) logln(severityLevel severity, args ...interface{}) {
	if !logger.shouldPrint(severityLevel) {
		return
	}

	if logger.prefix != "" {
		args = prependString(logger.prefix+":", args...)
	}

	logger.BackingLogger.Println(args...)
}

func (logger *Logger) log(severityLevel severity, args ...interface{}) {
	logger.logln(severityLevel, args...)
}

// ---------------------------
// Specific API implementation

// Debugf prints to stdout a formatted debug-level logging message
func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.logf(DebugSeverity, format, args...)
}

// Infof prints to stdout a formatted info-level logging message
func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.logf(InfoSeverity, format, args...)
}

// Warnf prints to stdout a formatted warning-level logging message
func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.logf(WarnSeverity, format, args...)
}

// Errorf prints to stdout a formatted error-level logging message
func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.logf(ErrorSeverity, format, args...)
}

// Panicf prints to stdout a formatted panic-level logging message
func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.logf(PanicSeverity, format, args...)
}

// Debugln prints to stdout a debug-level logging message
func (logger *Logger) Debugln(args ...interface{}) {
	logger.logln(DebugSeverity, args...)
}

// Infoln prints to stdout a info-level logging message
func (logger *Logger) Infoln(args ...interface{}) {
	logger.logln(InfoSeverity, args...)
}

// Warnln prints to stdout a warning-level logging message
func (logger *Logger) Warnln(args ...interface{}) {
	logger.logln(WarnSeverity, args...)
}

// Errorln prints to stdout a error-level logging message
func (logger *Logger) Errorln(args ...interface{}) {
	logger.logln(ErrorSeverity, args...)
}

// Panicln prints to stdout a panic-level logging message
func (logger *Logger) Panicln(args ...interface{}) {
	logger.logln(PanicSeverity, args...)
}

// Debug prints to stdout a debug-level logging message. Alias of
// Debugln method.
func (logger *Logger) Debug(args ...interface{}) {
	logger.log(DebugSeverity, args...)
}

// Info prints to stdout a info-level logging message. Alias of
// Infoln method.
func (logger *Logger) Info(args ...interface{}) {
	logger.log(InfoSeverity, args...)
}

// Warn prints to stdout a warning-level logging message. Alias of
// Warnln method.
func (logger *Logger) Warn(args ...interface{}) {
	logger.log(WarnSeverity, args...)
}

// Error prints to stdout a error-level logging message. Alias of
// Errorn method.
func (logger *Logger) Error(args ...interface{}) {
	logger.log(ErrorSeverity, args...)
}

// Panic prints to stdout a panic-level logging message. Alias of
// Panicln method.
func (logger *Logger) Panic(args ...interface{}) {
	logger.log(PanicSeverity, args...)
}
