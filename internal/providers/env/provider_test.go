package env

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	// Set an environment variable for testing
	os.Setenv("TEST_ENV_VAR", "test_env_val")
	defer os.Unsetenv("TEST_ENV_VAR")

	testCases := []struct {
		desc        string
		expectedID  string
		expectedVal string
		expectedErr error
	}{
		{
			desc:        "GetValue returns the value of the environment variable",
			expectedID:  "TEST_ENV_VAR",
			expectedVal: "test_env_val",
		},
		{
			desc:        "GetValue returns an error if the environment variable is not found",
			expectedID:  "TEST_ENV_VAR_NOT_FOUND",
			expectedErr: fmt.Errorf("%s cannot find environment variable '%s'", "env", "TEST_ENV_VAR_NOT_FOUND"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			provider := &EnvironmentProvider{
				Name: "env",
			}
			assert.Equal(t, "env", provider.GetName())

			val, err := provider.GetValue(testCase.expectedID)
			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedVal, string(val))
			}
		})
	}
}
