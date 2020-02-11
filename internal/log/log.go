package log

import (
	"fmt"
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

// severityLevels is a mapping of all available severity levels to their printed
// values.
var severityLevels = map[severity]string{
	DebugSeverity: "DEBUG",
	InfoSeverity:  "INFO",
	WarnSeverity:  "WARN",
	ErrorSeverity: "ERROR",
	PanicSeverity: "PANIC",
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
	return logger.IsDebug || (severityLevel != DebugSeverity)
}

func prependString(prependString string, args ...interface{}) []interface{} {
	prependSlice := []interface{}{prependString}
	return append(prependSlice, args...)
}

// DebugEnabled returns if the debug logging should be displayed for a particular
// logger instance
func (logger *Logger) DebugEnabled() bool {
	return logger.IsDebug
}

// CopyWith creates a copy of the logger with the prefix and debug values
// overridden by the arguments.
func (logger *Logger) CopyWith(prefix string, isDebug bool) log_api.Logger {
	return NewWithOptions(
		logger.BackingLogger.Writer(),
		prefix,
		isDebug,
	)
}

// Prefix returns the prefix that will be prepended to all output messages
func (logger *Logger) Prefix() string {
	return logger.prefix
}

func severityPrefix(sev severity) string {
	return fmt.Sprintf("%-7s", "["+severityLevels[sev]+"]")
}

// ---------------------------
// Main logging methods that funnel all the info here

func (logger *Logger) logf(sev severity, format string, args ...interface{}) {
	if !logger.shouldPrint(sev) {
		return
	}

	if logger.prefix != "" {
		format = "%s: " + format
		args = prependString(logger.prefix, args...)
	}

	format = "%s " + format
	args = prependString(severityPrefix(sev), args...)

	logger.BackingLogger.Printf(format, args...)
}

func (logger *Logger) logln(sev severity, args ...interface{}) {
	if !logger.shouldPrint(sev) {
		return
	}

	if logger.prefix != "" {
		args = prependString(logger.prefix+":", args...)
	}

	args = prependString(severityPrefix(sev), args...)

	logger.BackingLogger.Println(args...)
}

func (logger *Logger) log(sev severity, args ...interface{}) {
	logger.logln(sev, args...)
}

// TODO: This duplication is quite hideous, and should be cleaned up by
//   delegating everything to stdlib logger in a more straightforward way.
func (logger *Logger) panicf(sev severity, format string, args ...interface{}) {
	if !logger.shouldPrint(sev) {
		return
	}

	if logger.prefix != "" {
		format = "%s: " + format
		args = prependString(logger.prefix, args...)
	}

	format = "%s " + format
	args = prependString(severityPrefix(sev), args...)

	logger.BackingLogger.Panicf(format, args...)
}

func (logger *Logger) panicln(sev severity, args ...interface{}) {
	if !logger.shouldPrint(sev) {
		return
	}

	if logger.prefix != "" {
		args = prependString(logger.prefix+":", args...)
	}

	args = prependString(severityPrefix(sev), args...)

	logger.BackingLogger.Panicln(args...)
}

func (logger *Logger) panic(sev severity, args ...interface{}) {
	logger.panicln(sev, args...)
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
	logger.panicf(PanicSeverity, format, args...)
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
	logger.panicln(PanicSeverity, args...)
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
	logger.panic(PanicSeverity, args...)
}
