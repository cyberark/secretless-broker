package sharedobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksumVerifier(t *testing.T) {
	t.Run("Checksum verifier", func(t *testing.T) {
		t.Run("Doesn't error if there's no plugins to check", func(t *testing.T) {
		})

		t.Run("Doesn't error if checksums match", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/no_error_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			assert.NoError(t, err)
		})

		t.Run("Returns file list if checksums match", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/no_error_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			pluginFiles, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			assert.NoError(t, err)

			assert.Len(t, pluginFiles, 3)
			assert.Equal(t, pluginFiles[0].Name(), "bar.txt")
			assert.Equal(t, pluginFiles[1].Name(), "baz.txt")
			assert.Equal(t, pluginFiles[2].Name(), "foo.txt")
		})

		t.Run("Doesn't error if checksums match and checksums file prepends folder", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/folder_prepended.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			assert.NoError(t, err)
		})

		t.Run("Returns error if checksums file doesn't exist", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/doesntexist.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: open ./testdata/checksum/doesntexist.txt: no such file or directory"
			assert.EqualError(t, err, errorMsg)
		})

		t.Run("Returns error if checksums file is misformatted", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/misformatted_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: checksum file contained a misformatted line: 'fooo bar baz'"
			assert.EqualError(t, err, errorMsg)
		})

		t.Run("Returns error if unknown plugin is found in plugin list", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/unknown_plugin_hashes.txt"
			pluginDir := "./testdata/checksum/unknown_plugin"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: plugin 'unknown.txt' not found in checksums file"
			assert.EqualError(t, err, errorMsg)
		})

		t.Run("Returns error if plugin checksum doesn't match", func(t *testing.T) {
			checksumsFile := "./testdata/checksum/checksum_mismatch_hashes.txt"
			pluginDir := "./testdata/checksum/checksum_mismatch"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: plugin 'testdata/checksum/checksum_mismatch/foo.txt' checksum " +
				"'b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c' did not match " +
				"the expected 'DEADBEEFd4a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c'"
			assert.EqualError(t, err, errorMsg)
		})
	})
}
