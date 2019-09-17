package mock

import (
	"github.com/stretchr/testify/mock"
)

type credentialRetrieverMock struct {
	mock.Mock
}

func (cr *credentialRetrieverMock) RetrieveCredentials() (bytes map[string][]byte, e error) {
	args := cr.Called()

	// check for nil because the mock package is unable type assert nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string][]byte), args.Error(1)
}

// NewCredentialRetriever creates a mock with the `RetrieveCredentials` method
// that matches the signature of the CredentialsRetriever func type
func NewCredentialRetriever() *credentialRetrieverMock {
	return new(credentialRetrieverMock)
}
