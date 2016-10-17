package filesys

import (
	"io"
	"os"
	"time"
)

type FileSystem interface {
	Open(string) (File, error)
	Create(string) (File, error)
	Mkdir(string, os.FileMode) error
	MkdirAll(string, os.FileMode) error
	Remove(string) error
	RemoveAll(string) error
	Rename(oldPath, newPath string) error
	Stat(string) (os.FileInfo, error)
	Chtimes(name string, atime time.Time, mtime time.Time) error
	IsNotExist(error) bool
	IsExist(error) bool
	IsPermission(error) bool
	ReadDir(string) ([]os.FileInfo, error)
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, os.FileMode) error
}

type File interface {
	io.Reader
	io.Writer
	io.Closer
	io.Seeker
	Stat() (os.FileInfo, error)
	Readdir(n int) ([]os.FileInfo, error)
}
