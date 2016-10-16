package fsutil

import (
	"bytes"
	"fmt"
	"path"

	"github.com/poppels/filesys"
)

// CreateStructure creates multiple files and directories and returns
// the first error it encounters, if any
func CreateStructure(fs filesys.FileSystem, files map[string][]byte, dirs []string) error {
	if files != nil {
		if err := PutFiles(fs, files); err != nil {
			return err
		}
	}
	if dirs != nil {
		if err := MkdirMany(fs, dirs); err != nil {
			return err
		}
	}
	return nil
}

// PutFiles creates many files in a single call
func PutFiles(fs filesys.FileSystem, files map[string][]byte) error {
	for filepath, data := range files {
		if err := PutFile(fs, filepath, data); err != nil {
			return err
		}
	}
	return nil
}

// PutFile is like fs.WriteFile but it also creates all missing directories
func PutFile(fs filesys.FileSystem, name string, data []byte) error {
	dir, _ := path.Split(name)
	if err := fs.MkdirAll(dir, 0777); err != nil {
		return err
	}
	return fs.WriteFile(name, data, 0666)
}

// MkdirMany is a convenience method that runs MkdirAll for multiple paths
func MkdirMany(fs filesys.FileSystem, paths []string) error {
	for _, p := range paths {
		if err := fs.MkdirAll(p, 0777); err != nil {
			return err
		}
	}
	return nil
}

// VerifyFileContent checks that a file exists and contains the expected content,
// otherwise an error is returned.
func VerifyFileContent(fs filesys.FileSystem, name string, data []byte) error {
	content, err := fs.ReadFile(name)
	if err != nil {
		return err
	}
	if !bytes.Equal(content, data) {
		if len(content) > 20 {
			content = []byte(string(content[:17]) + "...")
		}
		if len(data) > 20 {
			data = []byte(string(data[:17]) + "...")
		}
		return fmt.Errorf("Expected file content '%s', got '%s'", data, content)
	}
	return nil
}
