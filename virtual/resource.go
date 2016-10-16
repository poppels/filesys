package virtual

import (
	"os"
	"sort"
	"time"
)

type resource struct {
	data     []byte
	children map[string]*resource
	isDir    bool
	name     string
	modTime  time.Time
	parent   *resource
}

func makeFolder(name string, parent *resource) *resource {
	return &resource{
		data:     nil,
		children: map[string]*resource{},
		isDir:    true,
		name:     name,
		modTime:  time.Now(),
		parent:   parent}
}

func makeFile(name string, parent *resource, data []byte) *resource {
	return &resource{
		data:     data,
		children: nil,
		isDir:    false,
		modTime:  time.Now(),
		name:     name,
		parent:   parent}
}

func (r *resource) stat() os.FileInfo {
	var size int64
	mode := os.ModeDir | 0777
	if !r.isDir {
		size = int64(len(r.data))
		mode = 0666
	}
	return VirtualFileInfo{
		size:    size,
		modTime: r.modTime,
		name:    r.name,
		mode:    mode}
}

func (r *resource) open(flag int) *VirtualFileHandle {
	return &VirtualFileHandle{
		res:      r,
		position: 0,
		canRead:  flag&os.O_RDONLY == os.O_RDONLY || flag&os.O_RDWR == os.O_RDWR,
		canWrite: flag&os.O_WRONLY == os.O_WRONLY || flag&os.O_RDWR == os.O_RDWR,
		modified: false,
		closed:   false}
}

func (r *resource) readdir() []os.FileInfo {
	infos := make([]os.FileInfo, 0, len(r.children))
	for _, sub := range r.children {
		infos = append(infos, sub.stat())
	}
	sort.Sort(byName(infos))
	return infos
}
