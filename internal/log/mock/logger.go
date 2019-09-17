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

func (l *LoggerMock) Errorf(format string, args ...interface{}) {
	l.Called()
	l.ReceivedCall <- struct{}{}

	return
}

// NewLogger creates a mock that conforms to the Secretless Logger interface
func NewLogger() *LoggerMock {
	mock := new(LoggerMock)
	mock.ReceivedCall = make(chan struct{})
	return mock
}
