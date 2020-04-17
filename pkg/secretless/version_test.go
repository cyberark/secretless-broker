package secretless

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionIsPresent(t *testing.T) {
	assert.NotEmpty(t, Version, "Expected Version to be non-empty but got an empty value")
}

func TestTagIsPresent(t *testing.T) {
	assert.NotEmpty(t, Tag, "Expected Tag to be non-empty but got an empty value")
}

func TestVersionIsCorrectFormat(t *testing.T) {
	assert.Regexp(t, `^[0-9]+\.[0-9]+\.[0-9]+$`, Version,
		"Expected Version to be a SemVer string")
}

func TestFullVersionNameIsCorrectFormat(t *testing.T) {
	assert.Regexp(t, `^[0-9]+\.[0-9]+\.[0-9]+-[a-z0-9]+$`, FullVersionName,
		"FullVersionName should be a '<SemVer>-<alphanumeric>' string")
}

// For now we enforce just plain lowercase alphanumerics but we might
// want to expand this later
func TestTagIsCorrectFormat(t *testing.T) {
	assert.Regexp(t, `^[a-z0-9]+$`, Tag,
		"Tag shoud be a strict lowercase alphanumeric string")
}
