package mock

import (
	"github.com/stretchr/testify/mock"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

type loggerMock struct {
	mock.Mock
	log.Logger
	ReceivedCall chan struct{}
}

func (l *loggerMock) Errorf(format string, args ...interface{}) {
	l.Called()
	l.ReceivedCall <- struct{}{}

	return
}

// NewLogger creates a mock that conforms to the Secretless Logger interface
func NewLogger() *loggerMock {
	mock := new(loggerMock)
	mock.ReceivedCall = make(chan struct{})
	return mock
}
