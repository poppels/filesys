package virtual

import (
	"os"
	"time"
)

type virtualFile struct {
	data    []byte
	name    string
	modTime time.Time
	parent  *virtualFolder
}

func makeFile(name string, parent *virtualFolder, data []byte) *virtualFile {
	return &virtualFile{
		data:    data,
		modTime: time.Now(),
		name:    name,
		parent:  parent}
}

func (f *virtualFile) stat() os.FileInfo {
	return VirtualFileInfo{
		size:    int64(len(f.data)),
		modTime: f.modTime,
		name:    f.name,
		mode:    0666}
}

func (f *virtualFile) open(flag int) *VirtualFileHandle {
	return &VirtualFileHandle{
		file:     f,
		folder:   nil,
		isDir:    false,
		position: 0,
		canRead:  flag&os.O_RDONLY == os.O_RDONLY || flag&os.O_RDWR == os.O_RDWR,
		canWrite: flag&os.O_WRONLY == os.O_WRONLY || flag&os.O_RDWR == os.O_RDWR,
		modified: false,
		closed:   false}
}
