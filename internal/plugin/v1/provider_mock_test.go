package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockProvider_GetName(t *testing.T) {
	p := &MockProvider{}

	assert.Equal(t, p.GetName(), "mock-provider")
}

func TestMockProvider_GetValue(t *testing.T) {
	p := &MockProvider{}

	t.Run("Get value", func(t *testing.T) {
		val, err := p.GetValue("foo")
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, string(val), "foo_value")
	})

	t.Run("Reports error", func(t *testing.T) {
		val, err := p.GetValue("err_foo")
		assert.EqualError(t, err, "err_foo_value")
		assert.Nil(t, val)
	})
}

func TestMockProvider_GetValues(t *testing.T) {
	p := &MockProvider{}

	t.Run("Sequentially calls GetValue", func(t *testing.T) {
		_, _ = p.GetValues("a", "b", "c")

		assert.Equal(t, []string{"a", "b", "c"}, p.GetValueCallArgs)
	})

	t.Run("Returns global error", func(t *testing.T) {
		res, err := p.GetValues("a", "b", "global_err_example", "c")
		assert.EqualError(t, err, "global_err_example_value")
		assert.Nil(t, res)
	})
}
