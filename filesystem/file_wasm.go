//go:build js && wasm

package filesystem

import (
	"io"
	"io/fs"
	"os"
	"syscall/js"
	"time"
)

type opfsFile struct {
	handle js.Value // The SyncAccessHandle
	path   string
	cursor int64
}

func (f *opfsFile) Write(p []byte) (n int, err error) {
	uint8Array := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(uint8Array, p)

	// Use manual cursor for 'at' to ensure we know where we are
	written := f.handle.Call("write", uint8Array, js.ValueOf(map[string]interface{}{
		"at": f.cursor,
	})).Int()

	f.cursor += int64(written)
	return written, nil
}

func (f *opfsFile) WriteAt(p []byte, off int64) (n int, err error) {
	uint8Array := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(uint8Array, p)

	// Standard WriteAt does NOT move the main file cursor
	written := f.handle.Call("write", uint8Array, js.ValueOf(map[string]interface{}{
		"at": off,
	})).Int()

	return written, nil
}

func (f *opfsFile) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	uint8Array := js.Global().Get("Uint8Array").New(len(p))

	// Read from the manual cursor
	read := f.handle.Call("read", uint8Array, js.ValueOf(map[string]interface{}{
		"at": f.cursor,
	})).Int()

	if read == 0 {
		// Check if we are at the end of the file
		size := int64(f.handle.Call("getSize").Int())
		if f.cursor >= size {
			return 0, io.EOF
		}
		return 0, nil
	}

	js.CopyBytesToGo(p, uint8Array)
	f.cursor += int64(read)
	return read, nil
}

func (f *opfsFile) ReadAt(p []byte, off int64) (n int, err error) {
	uint8Array := js.Global().Get("Uint8Array").New(len(p))

	read := f.handle.Call("read", uint8Array, js.ValueOf(map[string]interface{}{
		"at": off,
	})).Int()

	if read == 0 {
		return 0, io.EOF
	}

	js.CopyBytesToGo(p, uint8Array)
	return read, nil
}

func (f *opfsFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.cursor = offset
	case io.SeekCurrent:
		f.cursor += offset
	case io.SeekEnd:
		size := int64(f.handle.Call("getSize").Int())
		f.cursor = size + offset
	}

	// Prevent negative cursor
	if f.cursor < 0 {
		f.cursor = 0
	}

	return f.cursor, nil
}

func (f *opfsFile) Name() string {
	return f.path
}

func (f *opfsFile) Close() error {
	f.handle.Call("close")
	return nil
}

func (f *opfsFile) Lock() error {
	return nil
}

func (f *opfsFile) Unlock() error {
	return nil
}

func (f *opfsFile) Truncate(size int64) error {
	// OPFS SyncAccessHandle provides a native truncate method
	f.handle.Call("truncate", size)
	return nil
}

// ----------------------------------------------------------

type opfsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (fi opfsFileInfo) Name() string       { return fi.name }
func (fi opfsFileInfo) Size() int64        { return fi.size }
func (fi opfsFileInfo) ModTime() time.Time { return fi.modTime }
func (fi opfsFileInfo) IsDir() bool        { return fi.isDir }

// Mode returns the file permissions.
// Git checks this to see if a file is executable or a directory.
func (fi opfsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0o755
	}
	return 0o644
}

// Sys can return the underlying data source, usually nil for virtual FS.
func (fi opfsFileInfo) Sys() any {
	return nil
}

func (fi opfsFileInfo) Type() fs.FileMode {
	// We call fi.Mode() which we already implemented
	// and extract just the type bits.
	return fi.Mode().Type()
}

func (fi opfsFileInfo) Info() (fs.FileInfo, error) {
	// This is required if you are implementing the fs.DirEntry interface
	return fi, nil
}
