package osfilesys

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/poppels/filesys"
)

type OsFileSystem struct {
}

func NewOsWrapper() filesys.FileSystem {
	return OsFileSystem{}
}

func (OsFileSystem) Open(name string) (filesys.File, error) {
	file, err := os.Open(name)
	return file, err
}

func (OsFileSystem) Create(name string) (filesys.File, error) {
	file, err := os.Create(name)
	return file, err
}

func (OsFileSystem) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (OsFileSystem) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func (OsFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (OsFileSystem) RemoveAll(name string) error {
	return os.RemoveAll(name)
}

func (OsFileSystem) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (OsFileSystem) Stat(name string) (os.FileInfo, error) {
	fi, err := os.Stat(name)
	return fi, err
}

func (OsFileSystem) Chtimes(name string, atime, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}

func (OsFileSystem) ReadDir(name string) ([]os.FileInfo, error) {
	infos, err := ioutil.ReadDir(name)
	return infos, err
}

func (OsFileSystem) ReadFile(name string) ([]byte, error) {
	data, err := ioutil.ReadFile(name)
	return data, err
}

func (OsFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(name, data, perm)
}
