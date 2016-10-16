package osfilesys

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/poppels/filesys"
)

type OsFile struct {
	file *os.File
}

type OsFileSystem struct {
}

func NewOsWrapper() filesys.FileSystem {
	return OsFileSystem{}
}

func (f *OsFile) Read(b []byte) (int, error) {
	n, err := f.file.Read(b)
	return n, err
}

func (f *OsFile) Write(b []byte) (int, error) {
	n, err := f.file.Write(b)
	return n, err
}

func (f *OsFile) Seek(offset int64, whence int) (int64, error) {
	n, err := f.file.Seek(offset, whence)
	return n, err
}

func (f *OsFile) Stat() (os.FileInfo, error) {
	fi, err := f.file.Stat()
	return fi, err
}

func (f *OsFile) Readdir(n int) ([]os.FileInfo, error) {
	infos, err := f.file.Readdir(n)
	return infos, err
}

func (f *OsFile) Close() error {
	return f.file.Close()
}

func (OsFileSystem) Open(name string) (filesys.File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return &OsFile{file}, nil
}

func (OsFileSystem) Create(name string) (filesys.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &OsFile{file}, nil
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
