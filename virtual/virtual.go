package virtual

import (
	"errors"
	"os"
	"path"
	"strings"
	"time"

	"github.com/poppels/filesys"
)

var (
	errClosed             = errors.New("already closed")
	errNotReadable        = errors.New("not readable")
	errNotWritable        = errors.New("not writable")
	errNotEmpty           = errors.New("not empty")
	errIsDirectory        = errors.New("is a directory")
	errNotDirectory       = errors.New("not a directory")
	errInvalidDestination = errors.New("invalid destination")
	errInvalidPath        = errors.New("invalid path")
	errNegativeSeek       = errors.New("negative position")
	errSeekOverflow       = errors.New("new position is too large")
)

type VirtualFileSystem struct {
	root       *resource
	currentDir *resource
}

func NewVirtualFilesys() *VirtualFileSystem {
	root := makeFolder("", nil)
	return &VirtualFileSystem{root: root, currentDir: root}
}

func (fs *VirtualFileSystem) Mkdir(name string, perm os.FileMode) error {
	dir, filename := path.Split(path.Clean(name))
	if filename == "" || filename == "." {
		return nil
	}
	parent, err := fs.getFolder(dir, false)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}
	if c, exists := parent.children[filename]; exists && !c.isDir {
		return errNotDirectory
	}
	parent.children[filename] = makeFolder(filename, parent)
	return nil
}

func (fs *VirtualFileSystem) MkdirAll(name string, perm os.FileMode) error {
	_, err := fs.getFolder(name, true)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}
	return nil
}

func (fs *VirtualFileSystem) Remove(name string) error {
	err := fs.removePath(name, false)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}
	return nil
}

func (fs *VirtualFileSystem) RemoveAll(name string) error {
	err := fs.removePath(name, true)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}
	return nil
}

func (fs *VirtualFileSystem) Rename(oldPath, newPath string) error {
	if oldPath == "" || newPath == "" {
		return &os.LinkError{"rename", oldPath, newPath, errInvalidPath}
	}

	sourceResource, err := fs.getResource(oldPath)
	if err != nil {
		return &os.LinkError{"rename", oldPath, newPath, err}
	}

	source := path.Clean(oldPath)
	target := path.Clean(newPath)
	if source == target {
		return nil
	}
	if sourceResource == fs.root || target == "/" {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrPermission}
	}
	// Cannot move folder into itself or one of its descendants
	if strings.HasPrefix(target, source+"/") {
		return &os.LinkError{"rename", oldPath, newPath, errInvalidDestination}
	}

	targetDir, targetName := path.Split(target)

	targetParent, err := fs.getFolder(targetDir, false)
	if err != nil {
		return &os.LinkError{"rename", oldPath, newPath, err}
	}

	// Cannot overwrite folder neither with file nor folder
	if c, found := targetParent.children[targetName]; found && (sourceResource.isDir || c.isDir) {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrExist}
	}

	delete(sourceResource.parent.children, sourceResource.name)
	sourceResource.name = targetName
	targetParent.children[targetName] = sourceResource
	return nil
}

func (fs *VirtualFileSystem) Stat(name string) (os.FileInfo, error) {
	r, err := fs.getResource(name)
	if err != nil {
		return nil, &os.PathError{"stat", name, err}
	}
	return r.stat(), nil
}

func (fs *VirtualFileSystem) Open(name string) (filesys.File, error) {
	r, err := fs.getResource(name)
	if err != nil {
		return nil, &os.PathError{"open", name, err}
	}
	return r.open(os.O_RDONLY), nil
}

func (fs *VirtualFileSystem) Create(name string) (filesys.File, error) {
	f, err := fs.createFile(name)
	if err != nil {
		return nil, &os.PathError{"create", name, err}
	}
	return f.open(os.O_RDWR), nil
}

func (fs *VirtualFileSystem) Chtimes(name string, atime, mtime time.Time) error {
	r, err := fs.getResource(name)
	if err != nil {
		return &os.PathError{"chtimes", name, err}
	}
	r.modTime = mtime
	return nil
}

func (fs *VirtualFileSystem) ReadDir(name string) ([]os.FileInfo, error) {
	if name == "" {
		return nil, errInvalidPath
	}
	folder, err := fs.getFolder(name, false)
	if err != nil {
		return nil, &os.PathError{"readdir", name, err}
	}
	return folder.readdir(), err
}

func (fs *VirtualFileSystem) ReadFile(name string) ([]byte, error) {
	r, err := fs.getResource(name)
	if err != nil {
		return nil, &os.PathError{"readfile", name, err}
	}
	if r.isDir {
		return nil, &os.PathError{"readfile", name, errIsDirectory}
	}
	clone := make([]byte, len(r.data))
	copy(clone, r.data)
	return clone, nil
}

func (fs *VirtualFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := fs.createFile(name)
	if err != nil {
		return &os.PathError{"writefile", name, err}
	}
	f.data = make([]byte, len(data))
	copy(f.data, data)
	return nil
}

func (fs *VirtualFileSystem) ChangeDir(name string) (*VirtualFileSystem, error) {
	if name == "" {
		return nil, errInvalidPath
	}
	f, err := fs.getFolder(name, false)
	if err != nil {
		return nil, &os.PathError{"cd", name, err}
	}
	return &VirtualFileSystem{root: fs.root, currentDir: f}, nil
}

func (fs *VirtualFileSystem) CurrentDir() string {
	if fs.currentDir == fs.root {
		return "/"
	}

	// Get length of final path before allocating it
	length := 0
	for current := fs.currentDir; current != fs.root; current = current.parent {
		length += len(current.name) + 1
	}

	path := make([]byte, length)
	end := length
	for current := fs.currentDir; current != fs.root; current = current.parent {
		start := end - len(current.name)
		copy(path[start:end], current.name)
		end = start - 1
		path[end] = '/'
	}

	return string(path)
}

func (fs *VirtualFileSystem) createFile(name string) (*resource, error) {
	if strings.HasSuffix(name, "/") {
		return nil, errInvalidPath
	}
	dir, filename := path.Split(path.Clean(name))
	if filename == "" || filename == "." {
		return nil, errInvalidPath
	}

	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return nil, err
	}

	if c, exists := folder.children[filename]; exists && c.isDir {
		return nil, errIsDirectory
	}

	file := makeFile(filename, folder, []byte{})
	folder.children[filename] = file
	return file, nil
}

func (fs *VirtualFileSystem) getFolder(name string, createMissing bool) (*resource, error) {
	var current *resource
	if strings.HasPrefix(name, "/") {
		current = fs.root
	} else {
		current = fs.currentDir
	}
	parts := strings.Split(path.Clean(name), "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if current.parent != nil {
				current = current.parent
			}
			continue
		}

		child, exists := current.children[part]
		if exists {
			if !child.isDir {
				return nil, os.ErrNotExist
			}
			current = child
			continue
		}
		if !createMissing {
			return nil, os.ErrNotExist
		}

		child = makeFolder(part, current)
		current.children[part] = child
		current = child
	}

	return current, nil
}

func (fs *VirtualFileSystem) getResource(name string) (*resource, error) {
	if name == "/" {
		return nil, os.ErrPermission
	}
	dir, filename := path.Split(path.Clean(name))
	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return folder, nil
	}
	forceDir := strings.HasSuffix(name, "/")
	child, exists := folder.children[filename]
	if !exists || (forceDir && !child.isDir) {
		return nil, os.ErrNotExist
	}
	return child, nil
}

func (fs *VirtualFileSystem) removePath(name string, recursive bool) error {
	r, err := fs.getResource(name)
	if err != nil {
		return err
	}
	if r == fs.root {
		return os.ErrPermission
	}
	if !recursive && r.isDir && len(r.children) > 0 {
		return errNotEmpty
	}
	delete(r.parent.children, r.name)
	return nil
}
