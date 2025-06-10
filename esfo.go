package esfo

import (
    "os"
    "sync"
    "time"
)

// FileInfo mimics os.FileInfo for export to Swift.
type FileInfo struct {
    Name    string
    Size    int64
    Mode    uint32
    ModTime int64 // Unix timestamp
    IsDir   bool
}

// DirEntry mimics os.DirEntry for export to Swift.
type DirEntry struct {
    Name  string
    IsDir bool
}

// TempFileResult bundles CreateTemp results for Swift.
type TempFileResult struct {
    Filename string
    Fd       int64
}

// fileHandle tracks os.File and Swift file descriptor.
type fileHandle struct {
    file    *os.File // Go file object
    swiftFD int64    // Swift file descriptor
    name    string   // File name
}

var (
    fileHandles = make(map[int64]*fileHandle)
    handleMutex sync.Mutex
    nextHandle  int64 = 1
)

// addFileHandle maps a Swift FD to an os.File.
func addFileHandle(f *os.File, swiftFD int64, name string) {
    handleMutex.Lock()
    fileHandles[swiftFD] = &fileHandle{file: f, swiftFD: swiftFD, name: name}
    handleMutex.Unlock()
}

// getFileHandle retrieves os.File for a Swift FD.
func getFileHandle(swiftFD int64) (*fileHandle, bool) {
    handleMutex.Lock()
    fh, ok := fileHandles[swiftFD]
    handleMutex.Unlock()
    return fh, ok
}

// removeFileHandle removes a file handle.
func removeFileHandle(swiftFD int64) {
    handleMutex.Lock()
    delete(fileHandles, swiftFD)
    handleMutex.Unlock()
}

// Callbacks for Swift to implement.
var (
    writeFileCallback  func(filename string, data []byte, perm uint32) error
    readFileCallback   func(filename string) ([]byte, error)
    openFileCallback   func(name string, flag int, perm uint32) (int64, error)
    createCallback     func(name string) (int64, error)
    closeCallback      func(fd int64) error
    readCallback       func(fd int64, count int) ([]byte, error)
    writeCallback      func(fd int64, data []byte) (int, error)
    writeAtCallback    func(fd int64, data []byte, offset int64) (int, error)
    seekCallback       func(fd int64, offset int64, whence int) (int64, error)
    syncCallback       func(fd int64) error
    removeCallback     func(name string) error
    mkdirCallback      func(name string, perm uint32) error
    mkdirAllCallback   func(name string, perm uint32) error
    statCallback       func(name string) (FileInfo, error)
    chmodCallback      func(name string, mode uint32) error
    renameCallback     func(oldpath, newpath string) error
    readDirCallback    func(name string) ([]DirEntry, error)
    createTempCallback func(dir, pattern string) (TempFileResult, error)
    removeAllCallback  func(path string) error
    readLinkCallback   func(name string) (string, error)
    mkdirTempCallback  func(dir, pattern string) (string, error)
)

// SetWriteFileCallback sets the callback for WriteFile.
func SetWriteFileCallback(cb func(filename string, data []byte, perm uint32) error) {
    writeFileCallback = cb
}

// SetReadFileCallback sets the callback for ReadFile.
func SetReadFileCallback(cb func(filename string) ([]byte, error)) {
    readFileCallback = cb
}

// SetOpenFileCallback sets the callback for OpenFile.
func SetOpenFileCallback(cb func(name string, flag int, perm uint32) (int64, error)) {
    openFileCallback = cb
}

// SetCreateCallback sets the callback for Create.
func SetCreateCallback(cb func(name string) (int64, error)) {
    createCallback = cb
}

// SetCloseCallback sets the callback for Close.
func SetCloseCallback(cb func(fd int64) error) {
    closeCallback = cb
}

// SetReadCallback sets the callback for Read.
func SetReadCallback(cb func(fd int64, count int) ([]byte, error)) {
    readCallback = cb
}

// SetWriteCallback sets the callback for Write.
func SetWriteCallback(cb func(fd int64, data []byte) (int, error)) {
    writeCallback = cb
}

// SetWriteAtCallback sets the callback for WriteAt.
func SetWriteAtCallback(cb func(fd int64, data []byte, offset int64) (int, error)) {
    writeAtCallback = cb
}

// SetSeekCallback sets the callback for Seek.
func SetSeekCallback(cb func(fd int64, offset int64, whence int) (int64, error)) {
    seekCallback = cb
}

// SetSyncCallback sets the callback for Sync.
func SetSyncCallback(cb func(fd int64) error) {
    syncCallback = cb
}

// SetRemoveCallback sets the callback for Remove.
func SetRemoveCallback(cb func(name string) error) {
    removeCallback = cb
}

// SetMkdirCallback sets the callback for Mkdir.
func SetMkdirCallback(cb func(name string, perm uint32) error) {
    mkdirCallback = cb
}

// SetMkdirAllCallback sets the callback for MkdirAll.
func SetMkdirAllCallback(cb func(name string, perm uint32) error) {
    mkdirAllCallback = cb
}

// SetStatCallback sets the callback for Stat.
func SetStatCallback(cb func(name string) (FileInfo, error)) {
    statCallback = cb
}

// SetChmodCallback sets the callback for Chmod.
func SetChmodCallback(cb func(name string, mode uint32) error) {
    chmodCallback = cb
}

// SetRenameCallback sets the callback for Rename.
func SetRenameCallback(cb func(oldpath, newpath string) error) {
    renameCallback = cb
}

// SetReadDirCallback sets the callback for ReadDir.
func SetReadDirCallback(cb func(name string) ([]DirEntry, error)) {
    readDirCallback = cb
}

// SetCreateTempCallback sets the callback for CreateTemp.
func SetCreateTempCallback(cb func(dir, pattern string) (TempFileResult, error)) {
    createTempCallback = cb
}

// SetRemoveAllCallback sets the callback for RemoveAll.
func SetRemoveAllCallback(cb func(path string) error) {
    removeAllCallback = cb
}

// SetReadLinkCallback sets the callback for ReadLink.
func SetReadLinkCallback(cb func(name string) (string, error)) {
    readLinkCallback = cb
}

// SetMkdirTempCallback sets the callback for MkdirTemp.
func SetMkdirTempCallback(cb func(dir, pattern string) (string, error)) {
    mkdirTempCallback = cb
}

// File is a wrapper for os.File to satisfy os.FileInfo interface.
type File struct {
    *os.File
    swiftFD int64
    name    string
}

// Stat returns file info for the wrapped file.
func (f *File) Stat() (os.FileInfo, error) {
    return f.File.Stat()
}

// WriteFile writes data to the named file.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
    if writeFileCallback != nil {
        return writeFileCallback(filename, data, uint32(perm))
    }
    return os.WriteFile(filename, data, perm)
}

// ReadFile reads the named file.
func ReadFile(filename string) ([]byte, error) {
    if readFileCallback != nil {
        return readFileCallback(filename)
    }
    return os.ReadFile(filename)
}

// OpenFile opens the named file with specified flag and perm.
func OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
    if openFileCallback != nil {
        swiftFD, err := openFileCallback(name, flag, uint32(perm))
        if err != nil {
            return nil, err
        }
        // Create a temporary os.File for Go compatibility
        f, err := os.OpenFile(name, flag, perm)
        if err != nil {
            return nil, err
        }
        addFileHandle(f, swiftFD, name)
        return &File{File: f, swiftFD: swiftFD, name: name}, nil
    }
    f, err := os.OpenFile(name, flag, perm)
    if err != nil {
        return nil, err
    }
    return &File{File: f, swiftFD: 0, name: name}, nil
}

// Create creates or truncates the named file.
func Create(name string) (*File, error) {
    if createCallback != nil {
        swiftFD, err := createCallback(name)
        if err != nil {
            return nil, err
        }
        f, err := os.Create(name)
        if err != nil {
            return nil, err
        }
        addFileHandle(f, swiftFD, name)
        return &File{File: f, swiftFD: swiftFD, name: name}, nil
    }
    f, err := os.Create(name)
    if err != nil {
        return nil, err
    }
    return &File{File: f, swiftFD: 0, name: name}, nil
}

// Close closes the file.
func Close(f *File) error {
    if closeCallback != nil && f.swiftFD != 0 {
        err := closeCallback(f.swiftFD)
        if err != nil {
            return err
        }
        removeFileHandle(f.swiftFD)
        return f.File.Close()
    }
    return f.File.Close()
}

// Read reads up to len(b) bytes from the file.
func (f *File) Read(b []byte) (int, error) {
    if readCallback != nil && f.swiftFD != 0 {
        data, err := readCallback(f.swiftFD, len(b))
        if err != nil {
            return 0, err
        }
        n := copy(b, data)
        return n, nil
    }
    return f.File.Read(b)
}

// Write writes len(b) bytes to the file.
func (f *File) Write(b []byte) (int, error) {
    if writeCallback != nil && f.swiftFD != 0 {
        return writeCallback(f.swiftFD, b)
    }
    return f.File.Write(b)
}

// WriteAt writes len(b) bytes to the file at offset.
func (f *File) WriteAt(b []byte, off int64) (int, error) {
    if writeAtCallback != nil && f.swiftFD != 0 {
        return writeAtCallback(f.swiftFD, b, off)
    }
    return f.File.WriteAt(b, off)
}

// Seek sets the offset for the next Read or Write.
func (f *File) Seek(offset int64, whence int) (int64, error) {
    if seekCallback != nil && f.swiftFD != 0 {
        return seekCallback(f.swiftFD, offset, whence)
    }
    return f.File.Seek(offset, whence)
}

// Sync commits the file's contents to stable storage.
func (f *File) Sync() error {
    if syncCallback != nil && f.swiftFD != 0 {
        return syncCallback(f.swiftFD)
    }
    return f.File.Sync()
}

// Remove removes the named file or directory.
func Remove(name string) error {
    if removeCallback != nil {
        return removeCallback(name)
    }
    return os.Remove(name)
}

// Mkdir creates a directory named path.
func Mkdir(name string, perm os.FileMode) error {
    if mkdirCallback != nil {
        return mkdirCallback(name, uint32(perm))
    }
    return os.Mkdir(name, perm)
}

// MkdirAll creates a directory named path and parents.
func MkdirAll(name string, perm os.FileMode) error {
    if mkdirAllCallback != nil {
        return mkdirAllCallback(name, uint32(perm))
    }
    return os.MkdirAll(name, perm)
}

// Stat returns file information.
func Stat(name string) (os.FileInfo, error) {
    if statCallback != nil {
        fi, err := statCallback(name)
        if err != nil {
            return nil, err
        }
        return &fileInfo{
            name:    fi.Name,
            size:    fi.Size,
            mode:    os.FileMode(fi.Mode),
            modTime: time.Unix(fi.ModTime, 0),
            isDir:   fi.IsDir,
        }, nil
    }
    return os.Stat(name)
}

// fileInfo implements os.FileInfo.
type fileInfo struct {
    name    string
    size    int64
    mode    os.FileMode
    modTime time.Time
    isDir   bool
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() interface{}   { return nil }

// Chmod changes the mode of the named file.
func Chmod(name string, mode os.FileMode) error {
    if chmodCallback != nil {
        return chmodCallback(name, uint32(mode))
    }
    return os.Chmod(name, mode)
}

// Rename renames (moves) oldpath to newpath.
func Rename(oldpath, newpath string) error {
    if renameCallback != nil {
        return renameCallback(oldpath, newpath)
    }
    return os.Rename(oldpath, newpath)
}

// ReadDir reads the named directory.
func ReadDir(name string) ([]os.DirEntry, error) {
    if readDirCallback != nil {
        entries, err := readDirCallback(name)
        if err != nil {
            return nil, err
        }
        result := make([]os.DirEntry, len(entries))
        for i, e := range entries {
            result[i] = &dirEntry{name: e.Name, isDir: e.IsDir}
        }
        return result, nil
    }
    return os.ReadDir(name)
}

// dirEntry implements os.DirEntry.
type dirEntry struct {
    name  string
    isDir bool
}

func (d *dirEntry) Name() string               { return d.name }
func (d *dirEntry) IsDir() bool                { return d.isDir }
func (d *dirEntry) Type() os.FileMode          { return 0 }
func (d *dirEntry) Info() (os.FileInfo, error) { return nil, nil }

// CreateTemp creates a temporary file.
func CreateTemp(dir, pattern string) (*File, error) {
    if createTempCallback != nil {
        result, err := createTempCallback(dir, pattern)
        if err != nil {
            return nil, err
        }
        f, err := os.Create(result.Filename)
        if err != nil {
            return nil, err
        }
        addFileHandle(f, result.Fd, result.Filename)
        return &File{File: f, swiftFD: result.Fd, name: result.Filename}, nil
    }
    f, err := os.CreateTemp(dir, pattern)
    if err != nil {
        return nil, err
    }
    return &File{File: f, swiftFD: 0, name: f.Name()}, nil
}

// RemoveAll removes path and its children.
func RemoveAll(path string) error {
    if removeAllCallback != nil {
        return removeAllCallback(path)
    }
    return os.RemoveAll(path)
}

// ReadLink returns the destination of the named symbolic link.
func ReadLink(name string) (string, error) {
    if readLinkCallback != nil {
        return readLinkCallback(name)
    }
    return os.Readlink(name)
}

// MkdirTemp creates a temporary directory.
func MkdirTemp(dir, pattern string) (string, error) {
    if mkdirTempCallback != nil {
        return mkdirTempCallback(dir, pattern)
    }
    return os.MkdirTemp(dir, pattern)
}