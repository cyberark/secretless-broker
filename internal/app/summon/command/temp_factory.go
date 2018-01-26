package command

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// DEVSHM is the location of a memory-mapped directory on Linux.
const DEVSHM = "/dev/shm"

// TempFactory creates new temp files, using heuristics to choose as secure a location
// as possible.
type TempFactory struct {
	path  string
	files []string
}

// NewTempFactory creates a new temporary file factory.
// defer Cleanup() if you want the files removed.
func NewTempFactory(path string) TempFactory {
	if path == "" {
		path = DefaultTempPath()
	}
	return TempFactory{path: path}
}

// DefaultTempPath is the path for temporary files.
// Returns DEVSHM if it exists and is a directory. The home dir of current user
// is the next preferred option. The final option is the OS temp dir.
func DefaultTempPath() string {
	fi, err := os.Stat(DEVSHM)
	if err == nil && fi.Mode().IsDir() {
		return DEVSHM
	}
	home, err := homedir.Dir()
	if err == nil {
		dir, _ := ioutil.TempDir(home, ".tmp")
		return dir
	}
	return os.TempDir()
}

// Push creates a temp file with given value. Returns the path.
func (tf *TempFactory) Push(value string) string {
	f, _ := ioutil.TempFile(tf.path, ".summon")
	defer f.Close()

	f.Write([]byte(value))
	name := f.Name()
	tf.files = append(tf.files, name)
	return name
}

// Cleanup removes the temporary files created with this factory.
func (tf *TempFactory) Cleanup() {
	for _, file := range tf.files {
		os.Remove(file)
	}
	// Also remove the tempdir if it's not DEVSHM
	if !strings.Contains(tf.path, DEVSHM) {
		os.Remove(tf.path)
	}
	tf = nil
}
