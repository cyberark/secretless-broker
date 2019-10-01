package sharedobj

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestChecksumVerifier(t *testing.T) {
	Convey("Checksum verifier", t, func() {
		Convey("Doesn't error if there's no plugins to check", func() {
		})

		Convey("Doesn't error if checksums match", func() {
			checksumsFile := "./testdata/checksum/no_error_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			So(err, ShouldBeNil)
		})

		Convey("Returns file list if checksums match", func() {
			checksumsFile := "./testdata/checksum/no_error_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			pluginFiles, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			So(err, ShouldBeNil)

			So(len(pluginFiles), ShouldEqual, 3)
			So(pluginFiles[0].Name(), ShouldEqual, "bar.txt")
			So(pluginFiles[1].Name(), ShouldEqual, "baz.txt")
			So(pluginFiles[2].Name(), ShouldEqual, "foo.txt")
		})

		Convey("Doesn't error if checksums match and checksums file prepends folder", func() {
			checksumsFile := "./testdata/checksum/folder_prepended.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			So(err, ShouldBeNil)
		})

		Convey("Returns error if checksums file doesn't exist", func() {
			checksumsFile := "./testdata/checksum/doesntexist.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: open ./testdata/checksum/doesntexist.txt: no such file or directory"
			So(err.Error(), ShouldEqual, errorMsg)
		})

		Convey("Returns error if checksums file is misformatted", func() {
			checksumsFile := "./testdata/checksum/misformatted_hashes.txt"
			pluginDir := "./testdata/checksum/no_error"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: checksum file contained a misformatted line: 'fooo bar baz'"
			So(err.Error(), ShouldEqual, errorMsg)
		})

		Convey("Returns error if unknown plugin is found in plugin list", func() {
			checksumsFile := "./testdata/checksum/unknown_plugin_hashes.txt"
			pluginDir := "./testdata/checksum/unknown_plugin"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: plugin 'unknown.txt' not found in checksums file"
			So(err.Error(), ShouldEqual, errorMsg)
		})

		Convey("Returns error if plugin checksum doesn't match", func() {
			checksumsFile := "./testdata/checksum/checksum_mismatch_hashes.txt"
			pluginDir := "./testdata/checksum/checksum_mismatch"

			_, err := VerifyPluginChecksums(pluginDir, checksumsFile)

			errorMsg := "ERROR: plugin 'testdata/checksum/checksum_mismatch/foo.txt' checksum " +
				"'b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c' did not match " +
				"the expected 'DEADBEEFd4a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c'"
			So(err.Error(), ShouldEqual, errorMsg)
		})
	})
}
