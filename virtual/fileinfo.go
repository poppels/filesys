package virtual

import (
	"os"
	"time"
)

type VirtualFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi VirtualFileInfo) Name() string       { return fi.name }
func (fi VirtualFileInfo) Size() int64        { return fi.size }
func (fi VirtualFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi VirtualFileInfo) ModTime() time.Time { return fi.modTime }
func (fi VirtualFileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi VirtualFileInfo) Sys() interface{}   { return nil }

// For sorting file infos by names
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
