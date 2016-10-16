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
	errNegativeSeek       = errors.New("negative position")
	errSeekOverflow       = errors.New("new position is too large")
)

type VirtualFileSystem struct {
	root       *virtualFolder
	currentDir *virtualFolder
}

func NewVirtualFilesys() *VirtualFileSystem {
	root := makeFolder("", nil)
	return &VirtualFileSystem{root: root, currentDir: root}
}

func (fs *VirtualFileSystem) Mkdir(name string, perm os.FileMode) error {
	if fs == nil {
		return &os.PathError{"mkdir", name, os.ErrInvalid}
	}

	dir, filename := path.Split(path.Clean(name))
	if filename == "" { // Only happens when path is '/'
		return nil
	}

	parent, err := fs.getFolder(dir, false)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}

	if _, isFile := parent.files[filename]; isFile {
		return errNotDirectory
	}

	parent.folders[filename] = makeFolder(filename, parent)
	return nil
}

func (fs *VirtualFileSystem) MkdirAll(name string, perm os.FileMode) error {
	if fs == nil {
		return &os.PathError{"mkdir", name, os.ErrInvalid}
	}

	_, err := fs.getFolder(name, true)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}
	return nil
}

func (fs *VirtualFileSystem) Remove(name string) error {
	if fs == nil {
		return &os.PathError{"remove", name, os.ErrInvalid}
	}

	err := fs.removePath(name, false)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}

	return nil
}

func (fs *VirtualFileSystem) RemoveAll(name string) error {
	if fs == nil {
		return &os.PathError{"remove", name, os.ErrInvalid}
	}

	err := fs.removePath(name, true)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}

	return nil
}

func (fs *VirtualFileSystem) Rename(oldPath, newPath string) error {
	if fs == nil {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrInvalid}
	}
	if oldPath == "" || newPath == "" {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrNotExist}
	}

	source := path.Clean(oldPath)
	target := path.Clean(newPath)
	if source == target {
		return nil
	}
	if source == "/" || target == "/" {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrPermission}
	}

	sourceDir, sourceName := path.Split(source)
	targetDir, targetName := path.Split(target)

	sourceParent, err := fs.getFolder(sourceDir, false)
	if err != nil {
		return &os.LinkError{"rename", oldPath, newPath, err}
	}

	targetParent, err := fs.getFolder(targetDir, false)
	if err != nil {
		return &os.LinkError{"rename", oldPath, newPath, err}
	}

	sourceFolder, found := sourceParent.folders[sourceName]
	var sourceFile *virtualFile
	if !found {
		sourceFile, found = sourceParent.files[sourceName]
		if !found {
			return &os.LinkError{"rename", oldPath, newPath, os.ErrNotExist}
		}
	}

	// Cannot overwrite folder neither with file nor folder
	if _, found = targetParent.folders[targetName]; found {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrExist}
	}

	// Cannot overwrite a file with a folder
	if _, found = targetParent.files[targetName]; found && sourceFolder != nil {
		return &os.LinkError{"rename", oldPath, newPath, os.ErrExist}
	}

	if sourceFolder != nil {
		// Cannot move folder into itself or one of its descendants
		if strings.HasPrefix(target, source+"/") {
			return &os.LinkError{"rename", oldPath, newPath, errInvalidDestination}
		}

		delete(sourceParent.folders, sourceName)
		sourceFolder.name = targetName
		targetParent.folders[targetName] = sourceFolder
	} else {
		delete(sourceParent.files, sourceName)
		sourceFile.name = targetName
		targetParent.files[targetName] = sourceFile
	}

	return nil
}

func (fs *VirtualFileSystem) Stat(name string) (os.FileInfo, error) {
	if fs == nil {
		return nil, &os.PathError{"stat", name, os.ErrInvalid}
	}
	if name == "" {
		return nil, &os.PathError{"stat", name, os.ErrNotExist}
	}

	forceDir := strings.HasSuffix(name, "/")
	cleanPath := path.Clean(name)
	dir, filename := path.Split(cleanPath)
	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return nil, &os.PathError{"stat", name, err}
	}

	if file, ok := folder.files[filename]; ok {
		if forceDir {
			return nil, &os.PathError{"stat", name, errNotDirectory}
		}
		return file.stat(), nil
	}

	if sub, ok := folder.folders[filename]; ok {
		return sub.stat(), nil
	}

	return nil, &os.PathError{"stat", name, os.ErrNotExist}
}

func (fs *VirtualFileSystem) Open(name string) (filesys.File, error) {
	if fs == nil {
		return nil, &os.PathError{"open", name, os.ErrInvalid}
	}
	if name == "" {
		return nil, &os.PathError{"open", name, os.ErrNotExist}
	}

	forceDir := strings.HasSuffix(name, "/")
	cleanPath := path.Clean(name)
	dir, filename := path.Split(cleanPath)
	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return nil, &os.PathError{"open", name, err}
	}

	if !forceDir {
		if f, ok := folder.files[filename]; ok {
			return f.open(os.O_RDONLY), nil
		}
	}

	if f, ok := folder.folders[filename]; ok {
		return f.open(), nil
	}

	return nil, &os.PathError{"open", name, os.ErrNotExist}
}

func (fs *VirtualFileSystem) Create(name string) (filesys.File, error) {
	if fs == nil {
		return nil, &os.PathError{"create", name, os.ErrInvalid}
	}

	f, err := fs.createFile(name, false)
	if err != nil {
		return nil, &os.PathError{"create", name, err}
	}

	return f.open(os.O_RDWR), nil
}

func (fs *VirtualFileSystem) Chtimes(name string, atime, mtime time.Time) error {
	if fs == nil {
		return &os.PathError{"chtimes", name, os.ErrInvalid}
	}

	f, err := fs.getFile(name)
	if err != nil && !os.IsNotExist(err) {
		return &os.PathError{"chtimes", name, err}
	}

	if f != nil {
		f.modTime = mtime
		return nil
	}

	folder, err := fs.getFolder(name, false)
	if err != nil {
		return &os.PathError{"chtimes", name, err}
	}

	folder.modTime = mtime
	return nil
}

func (fs *VirtualFileSystem) ReadDir(name string) ([]os.FileInfo, error) {
	if fs == nil {
		return nil, &os.PathError{"readdir", name, os.ErrInvalid}
	}

	folder, err := fs.getFolder(name, false)
	if err != nil {
		return nil, &os.PathError{"readdir", name, err}
	}

	return folder.readdir(), nil
}

func (fs *VirtualFileSystem) ReadFile(name string) ([]byte, error) {
	if fs == nil {
		return nil, &os.PathError{"readfile", name, os.ErrInvalid}
	}

	f, err := fs.getFile(name)
	if err != nil {
		return nil, &os.PathError{"readfile", name, err}
	}

	return f.data, nil
}

func (fs *VirtualFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if fs == nil {
		return &os.PathError{"writefile", name, os.ErrInvalid}
	}

	f, err := fs.createFile(name, false)
	if err != nil {
		return &os.PathError{"writefile", name, err}
	}

	f.data = data
	return nil
}

func (fs *VirtualFileSystem) ChangeDir(name string) (*VirtualFileSystem, error) {
	if fs == nil {
		return nil, &os.PathError{"cd", name, os.ErrInvalid}
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
	for currentDir := fs.currentDir; currentDir != fs.root; currentDir = currentDir.parent {
		length += len(currentDir.name) + 1
	}

	path := make([]byte, length)
	end := length
	for currentDir := fs.currentDir; currentDir != fs.root; currentDir = currentDir.parent {
		start := end - len(currentDir.name)
		copy(path[start:end], currentDir.name)
		end = start - 1
		path[end] = '/'
	}

	return string(path)
}

// PutFile is like WriteFile but it also creates all missing directories
func (fs *VirtualFileSystem) PutFile(name string, content []byte) error {
	if fs == nil {
		return &os.PathError{"putfile", name, os.ErrInvalid}
	}

	f, err := fs.createFile(name, true)
	if err != nil {
		return &os.PathError{"putfile", name, err}
	}

	f.data = content
	return nil
}

func (fs *VirtualFileSystem) createFile(name string, createMissing bool) (*virtualFile, error) {
	if strings.HasSuffix(name, "/") {
		return nil, errInvalidDestination
	}
	dir, filename := path.Split(path.Clean(name))
	if filename == "" {
		return nil, errInvalidDestination
	}

	folder, err := fs.getFolder(dir, createMissing)
	if err != nil {
		return nil, err
	}

	if _, isDir := folder.folders[filename]; isDir {
		return nil, errIsDirectory
	}

	file := makeFile(filename, folder, []byte{})
	folder.files[filename] = file
	return file, nil
}

func (fs *VirtualFileSystem) getFolder(name string, createMissing bool) (*virtualFolder, error) {
	var currentDir *virtualFolder
	if strings.HasPrefix(name, "/") {
		currentDir = fs.root
	} else {
		currentDir = fs.currentDir
	}
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		if part == ".." {
			if currentDir.parent != nil {
				currentDir = currentDir.parent
			}
			continue
		}

		var exists bool
		folder, exists := currentDir.folders[part]
		if exists {
			currentDir = folder
			continue
		} else if !createMissing {
			return nil, os.ErrNotExist
		}

		if _, isFile := currentDir.files[part]; isFile {
			return nil, errNotDirectory
		}

		folder = makeFolder(part, currentDir)
		currentDir.folders[part] = folder
		currentDir = folder
	}

	return currentDir, nil
}

func (fs *VirtualFileSystem) getFile(name string) (*virtualFile, error) {
	dir, filename := path.Split(name)
	if filename == "" {
		return nil, os.ErrNotExist
	}

	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return nil, err
	}

	if f, ok := folder.files[filename]; ok {
		return f, nil
	}

	return nil, os.ErrNotExist
}

func (fs *VirtualFileSystem) removePath(name string, recursive bool) error {
	if name == "" {
		return os.ErrNotExist
	}

	forceDir := strings.HasSuffix(name, "/")
	cleanPath := path.Clean(name)
	dir, filename := path.Split(cleanPath)
	folder, err := fs.getFolder(dir, false)
	if err != nil {
		return err
	}

	if _, ok := folder.files[filename]; ok {
		if forceDir {
			return errNotDirectory
		}
		delete(folder.files, filename)
		return nil
	}

	if sub, ok := folder.folders[filename]; ok {
		if !recursive && len(sub.files)+len(sub.folders) > 0 {
			return errNotEmpty
		}
		delete(folder.folders, filename)
		return nil
	}

	return os.ErrNotExist
}
