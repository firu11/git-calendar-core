//go:build js && wasm

package filesystem

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"syscall/js"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/helper/chroot"
)

func GetRepoFS() (billy.Filesystem, error) {
	rootHandle := js.Global().Get("opfsRootHandle")

	return &OPFSStore{
		root: rootHandle,
		path: "/",
	}, nil
}

type OPFSStore struct {
	root js.Value // The OPFS DirectoryHandle
	path string   // Current working directory (for Root())
}

func (s *OPFSStore) MkdirAll(path string, perm fs.FileMode) error {
	// Tell JS to prepare this directory path in OPFS
	// We use a Promise-returning call but don't 'await' it here
	// because go-git assumes success if no error is returned.
	js.Global().Call("prepareDirectory", path)
	return nil
}

func (s *OPFSStore) Join(elem ...string) string {
	// Manual join to avoid path/filepath issues in some WASM envs
	res := ""
	for _, e := range elem {
		res += "/" + e
	}
	return res
}

func (s *OPFSStore) OpenFile(filename string, flag int, perm fs.FileMode) (billy.File, error) {
	// In your current code, this returns a Promise
	// You may need to use a helper that checks if the handle is ready
	handle := js.Global().Call("getHandleFromBucket", filename)
	if handle.IsUndefined() || handle.IsNull() {
		return nil, fs.ErrNotExist
	}
	return &opfsFile{handle: handle, path: filename}, nil
}

func (s *OPFSStore) Remove(filename string) error {
	s.root.Call("removeEntry", filename)
	return nil
}

func (s *OPFSStore) Rename(oldpath, newpath string) error {
	// OPFS doesn't have a native Rename yet in all browsers.
	// Common workaround: Move/Copy then Remove.
	fileHandle := s.root.Call("getFileHandle", oldpath)
	fileHandle.Call("move", newpath)
	return nil
}

func (s *OPFSStore) Root() string {
	return s.path
}

func (s *OPFSStore) Chroot(path string) (billy.Filesystem, error) {
	return chroot.New(s, path), nil
}

func (s *OPFSStore) ReadDir(path string) ([]os.FileInfo, error) {
	// Use a JS helper to get all names in the directory
	// You should define 'listEntries' in your worker.js
	jsEntries := js.Global().Call("listEntries", s.root, path)

	if jsEntries.IsUndefined() || jsEntries.IsNull() {
		return nil, fmt.Errorf("directory not found: %s", path)
	}

	length := jsEntries.Length()
	var entries []fs.FileInfo

	for i := range length {
		item := jsEntries.Index(i)
		entries = append(entries, &opfsFileInfo{
			name:  item.Get("name").String(),
			isDir: item.Get("kind").String() == "directory",
			// Size/ModTime can be fetched via Stat if needed,
			// but Git mostly just needs names from ReadDir
		})
	}

	return entries, nil
}

func (s *OPFSStore) Lstat(filename string) (fs.FileInfo, error) {
	return s.Stat(filename)
}

func (s *OPFSStore) Symlink(target, link string) error {
	return billy.ErrNotSupported
}

func (s *OPFSStore) Readlink(link string) (string, error) {
	return "", billy.ErrNotSupported
}

func (s *OPFSStore) TempFile(dir string, prefix string) (billy.File, error) {
	// Generate a unique filename: prefix + timestamp + random
	tempName := fmt.Sprintf("%s%d%d", prefix, time.Now().UnixNano(), rand.Intn(1000))
	fullPath := filepath.Join(dir, tempName)

	// Ensure the temp directory exists
	if dir != "" && dir != "." {
		_ = s.MkdirAll(dir, 0o755)
	}

	// Use your existing Create method to get a billy.File (opfsFile)
	return s.Create(fullPath)
}

func (s *OPFSStore) Create(filename string) (billy.File, error) {
	handle, err := s.callBridge(filename, "create")
	if err != nil {
		return nil, err
	}

	// WRAP the js.Value handle in your opfsFile struct
	return &opfsFile{
		handle: handle,
		path:   filename,
	}, nil
}

func (s *OPFSStore) Open(filename string) (billy.File, error) {
	handle, err := s.callBridge(filename, "open")
	if err != nil {
		return nil, err
	}

	// WRAP the js.Value handle in your opfsFile struct
	return &opfsFile{
		handle: handle,
		path:   filename,
	}, nil
}

// Internal bridge helper stays the same
func (s *OPFSStore) callBridge(path string, mode string) (js.Value, error) {
	ch := make(chan js.Value)
	cb := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- args[0]
		return nil
	})
	defer cb.Release()

	js.Global().Call("opfsBridge", path, mode, cb)
	result := <-ch

	if result.IsNull() || result.IsUndefined() {
		return js.Null(), fs.ErrNotExist
	}
	return result, nil
}

func (s *OPFSStore) Stat(path string) (fs.FileInfo, error) {
	res, err := s.callBridge(path, "stat")
	if err != nil {
		return nil, fs.ErrNotExist
	}
	return &opfsFileInfo{
		name:  path,
		isDir: res.Get("isDir").Bool(),
		size:  int64(res.Get("size").Int()),
	}, nil
}
