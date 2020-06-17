package ssl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDbSSLMode(t *testing.T) {
	t.Run("Options are passed as is", func(t *testing.T) {
		opts := options{
			"a": "b",
			"x": "y",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, sslmode.Options, opts)
	})

	t.Run("Invalid sslmode option", func(t *testing.T) {
		opts := options{
			"sslmode": "invalid",
		}

		_, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.Error(t, err) {
			return
		}
	})

	t.Run("sslmode=disable", func(t *testing.T) {
		opts := options{
			"sslmode": "disable",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.False(t, sslmode.UseTLS)
	})

	t.Run("sslmode=require", func(t *testing.T) {
		opts := options{
			"sslmode": "require",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.True(t, sslmode.UseTLS)
		assert.False(t, sslmode.VerifyCaOnly)
	})

	t.Run("sslmode=verify-ca", func(t *testing.T) {
		opts := options{
			"sslmode": "verify-ca",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.True(t, sslmode.UseTLS)
		assert.True(t, sslmode.VerifyCaOnly)
	})

	t.Run("sslmode=verify-full", func(t *testing.T) {
		opts := options{
			"sslmode": "verify-full",
			"host": "some-host",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.True(t, sslmode.UseTLS)
		assert.Equal(t, sslmode.ServerName, "some-host")
	})

	t.Run("sslmode=verify-full sslhost takes precedence", func(t *testing.T) {
		opts := options{
			"sslmode": "verify-full",
			"host": "some-host",
			"sslhost": "overridden-host",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.True(t, sslmode.UseTLS)
		assert.Equal(t, sslmode.ServerName, "overridden-host")
	})
}
