package sharedobj

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// VerifyPluginChecksums verifies all plugin files, and returns the FileInfo
// for the verified files.
func VerifyPluginChecksums(pluginDir string, checksumsFile string) ([]os.FileInfo, error) {
	log.Println("Verifying checksums of plugins...")

	pluginFiles, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("ERROR: %s", err)
	}

	checksums, err := loadChecksumsFile(checksumsFile)
	if err != nil {
		return nil, fmt.Errorf("ERROR: %s", err)
	}

	if err := compareChecksums(pluginDir, pluginFiles, checksums); err != nil {
		return nil, fmt.Errorf("ERROR: %s", err)
	}

	log.Println("Plugin verification completed.")
	return pluginFiles, nil
}

func compareChecksums(pluginDir string, pluginFiles []os.FileInfo, checksums map[string]string) error {
	for pluginIndex, pluginFile := range pluginFiles {
		pluginBasename := pluginFile.Name()
		fullPluginPath := path.Join(pluginDir, pluginBasename)

		actualChecksum, err := getSha256Sum(fullPluginPath)
		if err != nil {
			return err
		}

		log.Printf("- Plugin checksum verification (%d/%d): %s %s", pluginIndex+1, len(pluginFiles),
			actualChecksum, pluginBasename)

		expectedChecksum, ok := checksums[pluginBasename]
		if !ok {
			return fmt.Errorf("plugin '%s' not found in checksums file", pluginBasename)
		}

		if expectedChecksum != actualChecksum {
			return fmt.Errorf("plugin '%s' checksum '%s' did not match the expected '%s'",
				fullPluginPath, actualChecksum, expectedChecksum)
		}
	}

	return nil
}

func loadChecksumsFile(checksumsPath string) (map[string]string, error) {
	checksumsFile, err := os.Open(checksumsPath)
	if err != nil {
		return nil, err
	}
	defer checksumsFile.Close()

	checksumMap := map[string]string{}

	scanner := bufio.NewScanner(checksumsFile)
	for scanner.Scan() {
		checksumsLine := scanner.Text()

		fields := strings.Fields(checksumsLine)
		if len(fields) != 2 {
			formattingError := fmt.Errorf("checksum file contained a misformatted line: '%s'",
				checksumsLine)
			return nil, formattingError
		}

		checksum := fields[0]
		filename := filepath.Base(fields[1])

		checksumMap[filename] = checksum
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return checksumMap, nil
}

func getSha256Sum(filename string) (string, error) {
	filePt, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer filePt.Close()

	hashCalculator := sha256.New()
	if _, err := io.Copy(hashCalculator, filePt); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hashCalculator.Sum(nil)), nil
}
