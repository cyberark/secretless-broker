package mock

import (
	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

type loggerMock struct {
	mock.Mock
	log.Logger
}

func (l *loggerMock) Debugf(format string, args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Infof(format string, args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Warnf(format string, args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Errorf(format string, args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Panicf(format string, args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Debugln(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Infoln(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Warnln(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Errorln(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Panicln(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Debug(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Info(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Warn(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Error(args ...interface{}) {
	l.Called()
	return
}

func (l *loggerMock) Panic(args ...interface{}) {
	l.Called()
	return
}

// NewLogger creates a mock that conforms to the Secretless Logger interface
func NewLogger() *loggerMock {
	return new(loggerMock)
}
