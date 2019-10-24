package main

import (
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
)

func TestTemplateConnector(t *testing.T) {
	t.Run("An empty test", func(t *testing.T) {
		// We use go std test lib & testify/assert as our standard testing lib on secretless
	})
}
