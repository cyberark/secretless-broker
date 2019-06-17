package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
)

func TestGetStandardDeviation(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		input := &map[int]int{}
		res := GetStandardDeviation(input)

		assert.Equal(t, res, 0.0)
	})

	t.Run("empty input", func(t *testing.T) {
		input := &map[int]int{}
		res := GetStandardDeviation(input)

		assert.Equal(t, res, 0.0)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &map[int]int{
			1: 2,
			2: 3,
			3: 4,
		}
		res := GetStandardDeviation(input)

		// ((1 - 2.22...)^2 * 2 + (2 - 2.22...)^2 * 3 + (3 - 2.22...)^2 * 4) / (2 + 3 + 4 - 1))^0.5 = 0.833...
		assert.InDelta(t, res, 0.833, 0.01)
	})
}

func TestGetMean(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		input := &map[int]int{}
		res := GetMean(input)

		assert.Equal(t, res, 0.0)
	})

	t.Run("empty input", func(t *testing.T) {
		input := &map[int]int{}
		res := GetMean(input)

		assert.Equal(t, res, 0.0)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &map[int]int{
			1: 2,
			2: 3,
			3: 4,
		}
		res := GetMean(input)

		// (1 * 2 + 2 * 3 + 3 * 4) / (2 + 3 + 4) = 2.22...
		assert.InDelta(t, res, 2.22, 0.01)
	})
}

func TestGetAverageDuration(t *testing.T) {

	t.Run("empty input", func(t *testing.T) {
		input := &timing.BackendTiming{}
		res := GetAverageDuration(input)

		assert.Equal(t, res, time.Duration(0))
	})

	t.Run("nil input", func(t *testing.T) {
		input := &timing.BackendTiming{}
		res := GetAverageDuration(input)

		assert.Equal(t, res, time.Duration(0))
	})

	t.Run("valid input result is rounded down", func(t *testing.T) {
		input := &timing.BackendTiming{
			Count:    20,
			Duration: time.Duration(50),
			Errors:   make([]timing.TestRunError, 4),
		}
		res := GetAverageDuration(input)

		// 50 / (20 - 4) = 3.125
		assert.Equal(t, res, time.Duration(3))
	})
}

func Test_getMappedDataPointCount(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		input := &map[int]int{}
		res := getMappedDataPointCount(input)

		assert.Equal(t, res, 0)
	})

	t.Run("nil input", func(t *testing.T) {
		res := getMappedDataPointCount(nil)

		assert.Equal(t, res, 0)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &map[int]int{
			0: 1,
			5: 2,
			6: 3,
		}
		res := getMappedDataPointCount(input)

		// 1 + 2 + 3 = 6
		assert.Equal(t, res, 6)
	})
}
