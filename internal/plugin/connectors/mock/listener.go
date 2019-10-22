package mock

import (
	"net"

	"github.com/stretchr/testify/mock"
)

// ListenerMock conforms to the net.Listener interface
type ListenerMock struct {
	net.Listener
	mock.Mock
}

func numberOfMockMethodCalls(mock mock.Mock, method string) int {
	count := 0
	for _, call := range mock.Calls {
		if call.Method == method {
			count++
		}
	}

	return count
}

// Accept mocks the method of the same name on the net.Listener interface.
// It is a special mock method that normally blocks forever. When expected
// return values are set it will return those for the first call.
func (l *ListenerMock) Accept() (net.Conn, error) {
	args := l.Called()

	// block forever for calls that are not expected
	if numberOfMockMethodCalls(l.Mock, "Accept") > 1 {
		select {}
	}

	// check for nil because the mock package is unable type assert nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(net.Conn), args.Error(1)
}

// Close mocks the method of the same name on the net.Listener interface
func (l *ListenerMock) Close() error {
	args := l.Called()

	return args.Error(0)
}

// NewListener creates a net.Listener mock with an `Accept` method that returns
// the expectation values only on the first call, otherwise it blocks forever for
// all subsequent calls or if expected return values are not set.
func NewListener() *ListenerMock {
	return new(ListenerMock)
}
