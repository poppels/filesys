package virtual

import (
	"io"
	"os"
	"time"
)

type VirtualFileHandle struct {
	res      *resource
	position int
	canRead  bool
	canWrite bool
	modified bool
	closed   bool
}

func (fh *VirtualFileHandle) Read(b []byte) (int, error) {
	if fh.closed {
		return 0, &os.PathError{"read", fh.res.name, errClosed}
	}
	if fh.res.isDir {
		return 0, &os.PathError{"read", fh.res.name, errIsDirectory}
	}
	if !fh.canRead {
		return 0, &os.PathError{"read", fh.res.name, errNotReadable}
	}
	if len(b) == 0 {
		return 0, nil
	}
	if fh.position >= len(fh.res.data) {
		return 0, io.EOF
	}

	n := copy(b, fh.res.data[fh.position:])
	fh.position += n
	return n, nil
}

func (fh *VirtualFileHandle) Write(b []byte) (int, error) {
	if fh.closed {
		return 0, &os.PathError{"write", fh.res.name, errClosed}
	}
	if fh.res.isDir {
		return 0, &os.PathError{"write", fh.res.name, errIsDirectory}
	}
	if !fh.canWrite {
		return 0, &os.PathError{"write", fh.res.name, errNotWritable}
	}
	if len(b) == 0 {
		return 0, nil
	}

	// If position is after EOF, the gap should be filled with null bytes
	lenFile := len(fh.res.data)
	if fh.position > lenFile {
		filler := make([]byte, fh.position-lenFile)
		fh.res.data = append(fh.res.data, filler...)
		lenFile = fh.position
	}

	end := fh.position + len(b)
	if end < lenFile {
		copy(fh.res.data[fh.position:end], b)
	} else {
		fh.res.data = append(fh.res.data[:fh.position], b...)
	}

	fh.position = end
	fh.modified = true
	return len(b), nil
}

func (fh *VirtualFileHandle) Seek(offset int64, whence int) (int64, error) {
	if fh.closed {
		return 0, &os.PathError{"seek", fh.res.name, errClosed}
	}
	if fh.res.isDir {
		return 0, &os.PathError{"seek", fh.res.name, errIsDirectory}
	}
	if whence < 0 || whence > 2 {
		return 0, os.ErrInvalid
	}

	var pos int64
	if whence == io.SeekStart {
		pos = offset
	} else if whence == io.SeekCurrent {
		pos = int64(fh.position) + offset
	} else if whence == io.SeekEnd {
		pos = int64(len(fh.res.data)) + offset
	}

	if pos < 0 {
		return 0, &os.PathError{"seek", fh.res.name, errNegativeSeek}
	}
	if int64(int(pos)) != pos {
		return 0, &os.PathError{"seek", fh.res.name, errSeekOverflow}
	}

	fh.position = int(pos)
	return int64(fh.position), nil
}

func (fh *VirtualFileHandle) Stat() (os.FileInfo, error) {
	if fh.closed {
		return nil, &os.PathError{"stat", fh.res.name, errClosed}
	}
	return fh.res.stat(), nil
}

func (fh *VirtualFileHandle) Readdir(n int) ([]os.FileInfo, error) {
	if fh.closed {
		return nil, &os.PathError{"read", fh.res.name, errClosed}
	}
	if !fh.res.isDir {
		return nil, &os.PathError{"read", fh.res.name, errNotDirectory}
	}

	infos := fh.res.readdir()
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
	if !fh.res.isDir && fh.modified {
		fh.res.modTime = time.Now()
	}
	fh.closed = true
	return nil
}
