package virtual

import (
	"os"
	"sort"
	"time"
)

type virtualFolder struct {
	folders map[string]*virtualFolder
	files   map[string]*virtualFile
	name    string
	modTime time.Time
	parent  *virtualFolder
}

func makeFolder(name string, parent *virtualFolder) *virtualFolder {
	return &virtualFolder{
		folders: map[string]*virtualFolder{},
		files:   map[string]*virtualFile{},
		name:    name,
		modTime: time.Now(),
		parent:  parent}
}

func (f *virtualFolder) stat() os.FileInfo {
	return VirtualFileInfo{
		size:    0,
		modTime: f.modTime,
		name:    f.name,
		mode:    os.ModeDir | 0777}
}

func (f *virtualFolder) open() *VirtualFileHandle {
	return &VirtualFileHandle{
		file:     nil,
		folder:   f,
		isDir:    true,
		position: 0,
		canRead:  true,
		canWrite: false,
		modified: false,
		closed:   false}
}

func (f *virtualFolder) readdir() []os.FileInfo {
	infos := make([]os.FileInfo, 0, len(f.folders)+len(f.files))
	for _, sub := range f.folders {
		infos = append(infos, sub.stat())
	}
	for _, file := range f.files {
		infos = append(infos, file.stat())
	}
	sort.Sort(byName(infos))
	return infos
}
