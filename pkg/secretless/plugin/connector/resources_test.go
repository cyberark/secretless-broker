package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/log"
)

func TestNewResources(t *testing.T) {
	config := []byte("configvalue")
	logger := log.New(false)

	// Ensure that the return type is the exact type we expect
	var resources Resources

	resources = NewResources(config, logger)

	assert.Equal(t, config, resources.Config())
	assert.Equal(t, logger, resources.Logger())
}
