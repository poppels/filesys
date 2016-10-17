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
	Chtimes(path string, atime time.Time, mtime time.Time) error
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

var (
	singleton FileSystem
)

func SetGlobalSystem(fs FileSystem) {
	if fs == nil {
		panic("Argument is nil")
	}
	singleton = fs
}

func getSingleton() FileSystem {
	if singleton == nil {
		panic("Before any of the global filesys functions can be used you have to call filesys.SetGlobalSystem()")
	}
	return singleton
}

func Open(path string) (File, error) {
	fs := getSingleton()
	f, err := fs.Open(path)
	return f, err
}

func Create(path string) (File, error) {
	fs := getSingleton()
	f, err := fs.Create(path)
	return f, err
}

func Mkdir(path string, mode os.FileMode) error {
	fs := getSingleton()
	return fs.Mkdir(path, mode)
}

func MkdirAll(path string, mode os.FileMode) error {
	fs := getSingleton()
	return fs.MkdirAll(path, mode)
}

func Remove(path string) error {
	fs := getSingleton()
	return fs.Remove(path)
}

func RemoveAll(path string) error {
	fs := getSingleton()
	return fs.RemoveAll(path)
}

func Rename(oldPath, newPath string) error {
	fs := getSingleton()
	return fs.Rename(oldPath, newPath)
}

func Stat(path string) (os.FileInfo, error) {
	fs := getSingleton()
	fi, err := fs.Stat(path)
	return fi, err
}

func Chtimes(path string, atime time.Time, mtime time.Time) error {
	fs := getSingleton()
	return fs.Chtimes(path, atime, mtime)
}

func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func IsExist(err error) bool {
	return os.IsExist(err)
}

func IsPermission(err error) bool {
	return os.IsPermission(err)
}

func ReadDir(path string) ([]os.FileInfo, error) {
	fs := getSingleton()
	infos, err := fs.ReadDir(path)
	return infos, err
}

func ReadFile(path string) ([]byte, error) {
	fs := getSingleton()
	data, err := fs.ReadFile(path)
	return data, err
}

func WriteFile(path string, data []byte, mode os.FileMode) error {
	fs := getSingleton()
	return fs.WriteFile(path, data, mode)
}
