package virtual

import (
	"io"
	"os"
	"time"
)

type VirtualFileHandle struct {
	file     *virtualFile
	folder   *virtualFolder
	isDir    bool
	position int
	canRead  bool
	canWrite bool
	modified bool
	closed   bool
}

func (fh *VirtualFileHandle) Read(b []byte) (int, error) {
	if fh == nil {
		return 0, &os.PathError{"read", fh.file.name, os.ErrInvalid}
	}
	if fh.closed {
		return 0, &os.PathError{"read", fh.file.name, errClosed}
	}
	if fh.isDir {
		return 0, &os.PathError{"read", fh.file.name, errIsDirectory}
	}
	if !fh.canRead {
		return 0, &os.PathError{"read", fh.file.name, errNotReadable}
	}
	if len(b) == 0 {
		return 0, nil
	}
	if fh.position >= len(fh.file.data) {
		return 0, io.EOF
	}

	n := copy(b, fh.file.data[fh.position:])
	fh.position += n
	return n, nil
}

func (fh *VirtualFileHandle) Write(b []byte) (int, error) {
	if fh == nil {
		return 0, &os.PathError{"write", fh.file.name, os.ErrInvalid}
	}
	if fh.closed {
		return 0, &os.PathError{"write", fh.file.name, errClosed}
	}
	if fh.isDir {
		return 0, &os.PathError{"write", fh.file.name, errIsDirectory}
	}
	if !fh.canWrite {
		return 0, &os.PathError{"write", fh.file.name, errNotWritable}
	}
	if len(b) == 0 {
		return 0, nil
	}

	// If position is after EOF, the gap should be filled with null bytes
	lenFile := len(fh.file.data)
	if fh.position > lenFile {
		filler := make([]byte, fh.position-lenFile)
		fh.file.data = append(fh.file.data, filler...)
		lenFile = fh.position
	}

	end := fh.position + len(b)
	if end < lenFile {
		copy(fh.file.data[fh.position:end], b)
	} else {
		fh.file.data = append(fh.file.data[:fh.position], b...)
	}

	fh.position = end
	fh.modified = true
	return len(b), nil
}

func (fh *VirtualFileHandle) Seek(offset int64, whence int) (int64, error) {
	if fh == nil {
		return 0, &os.PathError{"seek", fh.file.name, os.ErrInvalid}
	}
	if fh.closed {
		return 0, &os.PathError{"seek", fh.file.name, errClosed}
	}
	if fh.isDir {
		return 0, &os.PathError{"seek", fh.file.name, errIsDirectory}
	}
	if fh == nil || whence < 0 || whence > 2 {
		return 0, os.ErrInvalid
	}

	var pos int64
	if whence == 0 {
		pos = offset
	} else if whence == 1 {
		pos = int64(fh.position) + offset
	} else if whence == 2 {
		pos = int64(len(fh.file.data)) + offset
	}

	if pos < 0 {
		return 0, &os.PathError{"seek", fh.file.name, errNegativeSeek}
	}
	if int64(int(pos)) != pos {
		return 0, &os.PathError{"seek", fh.file.name, errSeekOverflow}
	}

	fh.position = int(pos)
	return int64(fh.position), nil
}

func (fh *VirtualFileHandle) Stat() (os.FileInfo, error) {
	if fh == nil {
		return nil, &os.PathError{"stat", fh.file.name, os.ErrInvalid}
	}
	if fh.closed {
		return nil, &os.PathError{"stat", fh.file.name, errClosed}
	}
	if fh.isDir {
		return fh.folder.stat(), nil
	}
	return fh.file.stat(), nil
}

func (fh *VirtualFileHandle) Readdir(n int) ([]os.FileInfo, error) {
	if fh == nil {
		return nil, &os.PathError{"read", fh.file.name, os.ErrInvalid}
	}
	if fh.closed {
		return nil, &os.PathError{"read", fh.file.name, errClosed}
	}
	if !fh.isDir {
		return nil, &os.PathError{"read", fh.file.name, errNotDirectory}
	}

	infos := fh.folder.readdir()
	var err error
	if n <= 0 {
		infos = infos[fh.position:]
		fh.position = len(infos)
	} else {
		end := fh.position + n
		if end > len(infos) {
			err = io.EOF
			end = len(infos)
		}
		infos = infos[fh.position:end]
		fh.position = end
	}

	return infos, err
}

func (fh *VirtualFileHandle) Close() error {
	if fh == nil {
		return &os.PathError{"close", fh.file.name, os.ErrInvalid}
	}

	if !fh.isDir && fh.modified {
		fh.file.modTime = time.Now()
	}

	fh.closed = true
	return nil
}
