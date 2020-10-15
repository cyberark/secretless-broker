package command

import (
	"io/ioutil"
	"os"
	"strings"
)

// This is the location of a memory-mapped directory on Linux. We want
// to use this location where possible since files saved there are not
// saved on disk but kept in transient memory which is more secure than
// the filesystem.
const defaultSharedMemoryDir = "/dev/shm"

// TempFactory creates new temp files, using heuristics to choose as secure a location
// as possible.
type TempFactory struct {
	path  string
	files []string
}

// NewTempFactory creates a new temporary file factory.
// defer Cleanup() if you want the files removed.
func NewTempFactory(path string) TempFactory {
	return NewCustomTempFactory(path, "")
}

// NewCustomTempFactory creates a new temporary file factory with specified
// sharedMemoryDir. If sharedMemoryDir is empty, we use the default path for it.
// defer Cleanup() if you want the files removed.
func NewCustomTempFactory(path string, sharedMemoryDir string) TempFactory {
	if path == "" {
		path = defaultTempPath(sharedMemoryDir)
	}

	return TempFactory{
		path: path,
	}
}

// defaultTempPath is the path for temporary files.
// Returns shared memory dir if it exists and is a directory. The home dir of
// current user is the next preferred option. The final option is the OS temp dir.
func defaultTempPath(sharedMemoryDir string) string {
	if sharedMemoryDir == "" {
		sharedMemoryDir = defaultSharedMemoryDir
	}

	fi, err := os.Stat(sharedMemoryDir)
	if err == nil && fi.Mode().IsDir() {
		return sharedMemoryDir
	}

	home, err := os.UserHomeDir()
	if err == nil {
		dir, err := ioutil.TempDir(home, ".tmp")
		if err == nil {
			return dir
		}
	}

	return os.TempDir()
}

// Push creates a temp file with given value. Returns the path.
func (tf *TempFactory) Push(value string) (string, error) {
	f, err := ioutil.TempFile(tf.path, ".summon")
	if err != nil {
		return "", err
	}

	defer f.Close()

	_, err = f.Write([]byte(value))
	if err != nil {
		return "", err
	}

	name := f.Name()
	tf.files = append(tf.files, name)
	return name, nil
}

// Cleanup removes the temporary files created with this factory.
func (tf *TempFactory) Cleanup() {
	for _, file := range tf.files {
		_ = os.Remove(file)
	}
	tf.files = nil

	// Also remove the tempdir if it's not the shared memory directory
	if !strings.Contains(tf.path, defaultSharedMemoryDir) {
		_ = os.Remove(tf.path)
	}
}
