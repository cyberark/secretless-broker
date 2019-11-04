package profile_test

import (
	"testing"

	gh_profile "github.com/pkg/profile"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/profile"
)

// We cannot name/alias the Stop() interface since the type casting
// breaks and we can't change the underlying class.
type stopperFunc func()

func (fn stopperFunc) Stop() {
	fn()
}

func TestProfile(t *testing.T) {
	t.Run("New profile only allows cpu and memory types", func(t *testing.T) {
		for _, profileType := range []string{"cpu", "memory"} {
			profile, err := profile.New(profileType)

			assert.NotNil(t, profile)
			assert.Nil(t, err)
		}

		for _, profileType := range []string{"foo", "bar"} {
			profile, err := profile.New(profileType)

			assert.Nil(t, profile)

			// Sanity check
			assert.NotNil(t, err)
			if err == nil {
				continue
			}

			assert.Contains(t, err.Error(), "Invalid profile type")
			assert.Contains(t, err.Error(), profileType)
		}
	})

	t.Run("New profile delegates to profile interface object", func(t *testing.T) {
		startCalled := false
		stopCalled := false

		mockProfiler := func(profiles ...func(*gh_profile.Profile)) interface{ Stop() } {
			startCalled = true

			profilerStopper := func() {
				stopCalled = true
			}

			return stopperFunc(profilerStopper)
		}

		profile, err := profile.NewWithOptions("cpu", mockProfiler)

		// Sanity check
		assert.Nil(t, err)
		if err != nil {
			return
		}

		// Sanity check #2
		assert.False(t, startCalled)
		assert.False(t, stopCalled)

		profile.Start()

		assert.True(t, startCalled)
		assert.False(t, stopCalled)

		profile.Stop()

		assert.True(t, startCalled)
		assert.True(t, stopCalled)
	})
}
