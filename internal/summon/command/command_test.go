package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convertSubsToMap(t *testing.T) {
	expected := map[string]string{"foo": "bar=foo", "bar": "foo=bar"}
	actual := convertSubsToMap([]string{"foo=bar=foo", "bar=foo=bar"})

	assert.Equal(t, expected, actual)
}
