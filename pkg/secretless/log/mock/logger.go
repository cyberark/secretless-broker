package mock

import (
	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

// LoggerMock conforms to the Secretless Logger interface
type LoggerMock struct {
	mock.Mock
	log.Logger
	ReceivedCall chan struct{}
}

// Errorf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Errorf(format string, args ...interface{}) {
	l.Called()
	l.ReceivedCall <- struct{}{}

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
func (l *LoggerMock) Warn(...interface{}) {}

// Warnf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Warnf(string, ...interface{}) {}

// Warnln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Warnln(...interface{}) {}

// Error mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Error(...interface{}) {}

// Errorln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Errorln(...interface{}) {}

// Panic mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panic(...interface{}) {}

// Panicf mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panicf(string, ...interface{}) {}

// Panicln mocks the method of the same name on the log.Logger interface
func (l *LoggerMock) Panicln(...interface{}) {}

// NewLogger creates a mock that conforms to the Secretless Logger interface
func NewLogger() *LoggerMock {
	mock := new(LoggerMock)
	mock.ReceivedCall = make(chan struct{})
	return mock
}
