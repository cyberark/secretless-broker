package mock

import (
	"fmt"

	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

// LoggerMock conforms to the Secretless Logger interface
type LoggerMock struct {
	mock.Mock
	log.Logger
	ReceivedCall chan struct{}
	Warns        []string
	Errors       []string
	Panics       []string
}

// Errorf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Errorf(format string, args ...interface{}) {
	l.Called()
	l.ReceivedCall <- struct{}{}
	l.Errors = append(l.Errors, fmt.Sprintf(format, args...))

	return
}

// CopyWith mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) CopyWith(prefix string, isDebug bool) log.Logger {
	return new(LoggerMock)
}

// DebugEnabled mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) DebugEnabled() bool { return false }

// Prefix mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Prefix() string { return "" }

// Debug mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Debug(...interface{}) {}

// Debugf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Debugf(string, ...interface{}) {}

// Debugln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Debugln(...interface{}) {}

// Info mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Info(...interface{}) {}

// Infof mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Infof(string, ...interface{}) {}

// Infoln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Infoln(...interface{}) {}

// Warn mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Warn(args ...interface{}) {
	l.Warns = append(l.Warns, fmt.Sprint(args...))
}

// Warnf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Warnf(format string, args ...interface{}) {
	l.Warns = append(l.Warns, fmt.Sprintf(format, args...))
}

// Warnln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Warnln(args ...interface{}) {
	l.Warns = append(l.Warns, fmt.Sprintln(args...))
}

// Error mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Error(args ...interface{}) {
	l.Errors = append(l.Errors, fmt.Sprint(args...))
}

// Errorln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Errorln(args ...interface{}) {
	l.Errors = append(l.Errors, fmt.Sprintln(args...))
}

// Panic mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panic(args ...interface{}) {
	l.Panics = append(l.Panics, fmt.Sprint(args...))
}

// Panicf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panicf(format string, args ...interface{}) {
	l.Panics = append(l.Panics, fmt.Sprintf(format, args...))
}

// Panicln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panicln(args ...interface{}) {
	l.Panics = append(l.Panics, fmt.Sprintln(args...))
}

// NewLogger creates a mock that conforms to the Secretless Logger interface
func NewLogger() *LoggerMock {
	mock := new(LoggerMock)
	mock.ReceivedCall = make(chan struct{})
	return mock
}
