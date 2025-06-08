package esfo

import (
    "os"
    "sync"
    "time"
)

// TempFileResult bundles CreateTemp results.
type TempFileResult struct {
    Filename string
    Fd       int64
}

// FileInfo mimics os.FileInfo for GoMobile compatibility.
type FileInfo struct {
    Name    string
    Size    int64
    Mode    uint32
    ModTime int64 // Unix timestamp (seconds since epoch)
    IsDir   bool
}

// DirEntry mimics os.DirEntry for GoMobile compatibility.
type DirEntry struct {
    Name  string
    IsDir bool
}

// wrappedFile stores metadata for Swift file operations.
type wrappedFile struct {
    file    *os.File // Underlying os.File
    swiftFD int64    // File descriptor ID
    name    string   // File name for Swift
}

// swiftFileMap stores wrappedFile entries for Swift-managed files.
var (
    swiftFileMap = make(map[uintptr]*wrappedFile)
    mapMutex     sync.Mutex
)

// addSwiftFile associates a file with Swift metadata.
func addSwiftFile(f *os.File, swiftFD int64, name string) {
    mapMutex.Lock()
    swiftFileMap[f.Fd()] = &wrappedFile{file: f, swiftFD: swiftFD, name: name}
    mapMutex.Unlock()
}

// getSwiftFile retrieves Swift metadata for a file.
func getSwiftFile(f *os.File) (*wrappedFile, bool) {
    mapMutex.Lock()
    wf, ok := swiftFileMap[f.Fd()]
    mapMutex.Unlock()
    return wf, ok
}

// removeSwiftFile removes Swift metadata for a file.
func removeSwiftFile(f *os.File) {
    mapMutex.Lock()
    delete(swiftFileMap, f.Fd())
    mapMutex.Unlock()
}

// Callback functions for Swift to set.
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
func SetWriteFileCallback(filename string, data []byte, perm uint32) error {
    if writeFileCallback != nil {
        return writeFileCallback(filename, data, perm)
    }
    return os.ErrInvalid
}

// SetReadFileCallback sets the callback for ReadFile.
func SetReadFileCallback(filename string) ([]byte, error) {
    if readFileCallback != nil {
        return readFileCallback(filename)
    }
    return nil, os.ErrInvalid
}

// SetOpenFileCallback sets the callback for OpenFile.
func SetOpenFileCallback(name string, flag int, perm uint32) (int64, error) {
    if openFileCallback != nil {
        return openFileCallback(name, flag, perm)
    }
    return 0, os.ErrInvalid
}

// SetCreateCallback sets the callback for Create.
func SetCreateCallback(name string) (int64, error) {
    if createCallback != nil {
        return createCallback(name)
    }
    return 0, os.ErrInvalid
}

// SetCloseCallback sets the callback for Close.
func SetCloseCallback(fd int64) error {
    if closeCallback != nil {
        return closeCallback(fd)
    }
    return os.ErrInvalid
}

// SetReadCallback sets the callback for Read.
func SetReadCallback(fd int64, count int) ([]byte, error) {
    if readCallback != nil {
        return readCallback(fd, count)
    }
    return nil, os.ErrInvalid
}

// SetWriteCallback sets the callback for Write.
func SetWriteCallback(fd int64, data []byte) (int, error) {
    if writeCallback != nil {
        return writeCallback(fd, data)
    }
    return 0, os.ErrInvalid
}

// SetWriteAtCallback sets the callback for WriteAt.
func SetWriteAtCallback(fd int64, data []byte, offset int64) (int, error) {
    if writeAtCallback != nil {
        return writeAtCallback(fd, data, offset)
    }
    return 0, os.ErrInvalid
}

// SetSeekCallback sets the callback for Seek.
func SetSeekCallback(fd int64, offset int64, whence int) (int64, error) {
    if seekCallback != nil {
        return seekCallback(fd, offset, whence)
    }
    return 0, os.ErrInvalid
}

// SetSyncCallback sets the callback for Sync.
func SetSyncCallback(fd int64) error {
    if syncCallback != nil {
        return syncCallback(fd)
    }
    return os.ErrInvalid
}

// SetRemoveCallback sets the callback for Remove.
func SetRemoveCallback(name string) error {
    if removeCallback != nil {
        return removeCallback(name)
    }
    return os.ErrInvalid
}

// SetMkdirCallback sets the callback for Mkdir.
func SetMkdirCallback(name string, perm uint32) error {
    if mkdirCallback != nil {
        return mkdirCallback(name, perm)
    }
    return os.ErrInvalid
}

// SetMkdirAllCallback sets the callback for MkdirAll.
func SetMkdirAllCallback(name string, perm uint32) error {
    if mkdirAllCallback != nil {
        return mkdirAllCallback(name, perm)
    }
    return os.ErrInvalid
}

// SetStatCallback sets the callback for Stat.
func SetStatCallback(name string) (FileInfo, error) {
    if statCallback != nil {
        return statCallback(name)
    }
    return FileInfo{}, os.ErrInvalid
}

// SetChmodCallback sets the callback for Chmod.
func SetChmodCallback(name string, mode uint32) error {
    if chmodCallback != nil {
        return chmodCallback(name, mode)
    }
    return os.ErrInvalid
}

// SetRenameCallback sets the callback for Rename.
func SetRenameCallback(oldpath, newpath string) error {
    if renameCallback != nil {
        return renameCallback(oldpath, newpath)
    }
    return os.ErrInvalid
}

// SetReadDirCallback sets the callback for ReadDir.
func SetReadDirCallback(name string) ([]DirEntry, error) {
    if readDirCallback != nil {
        return readDirCallback(name)
    }
    return nil, os.ErrInvalid
}

// SetCreateTempCallback sets the callback for CreateTemp.
func SetCreateTempCallback(dir, pattern string) (TempFileResult, error) {
    if createTempCallback != nil {
        return createTempCallback(dir, pattern)
    }
    return TempFileResult{}, os.ErrInvalid
}

// SetRemoveAllCallback sets the callback for RemoveAll.
func SetRemoveAllCallback(path string) error {
    if removeAllCallback != nil {
        return removeAllCallback(path)
    }
    return os.ErrInvalid
}

// SetReadLinkCallback sets the callback for ReadLink.
func SetReadLinkCallback(name string) (string, error) {
    if readLinkCallback != nil {
        return readLinkCallback(name)
    }
    return "", os.ErrInvalid
}

// SetMkdirTempCallback sets the callback for MkdirTemp.
func SetMkdirTempCallback(dir, pattern string) (string, error) {
    if mkdirTempCallback != nil {
        return mkdirTempCallback(dir, pattern)
    }
    return "", os.ErrInvalid
}

// WriteFile writes data to the named file with the given permissions.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
    if writeFileCallback != nil {
        return writeFileCallback(filename, data, uint32(perm))
    }
    return os.WriteFile(filename, data, perm)
}

// ReadFile reads the named file and returns its contents.
func ReadFile(filename string) ([]byte, error) {
    if readFileCallback != nil {
        return readFileCallback(filename)
    }
    return os.ReadFile(filename)
}

// OpenFile opens the named file with specified flag and permissions.
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
    if openFileCallback != nil {
        fd, err := openFileCallback(name, flag, uint32(perm))
        if err != nil {
            return nil, err
        }
        f, err := os.OpenFile(name, flag, perm)
        if err != nil {
            closeCallback(fd)
            return nil, err
        }
        addSwiftFile(f, fd, name)
        return f, nil
    }
    return os.OpenFile(name, flag, perm)
}

// Create creates or truncates the named file.
func Create(name string) (*os.File, error) {
    if createCallback != nil {
        fd, err := createCallback(name)
        if err != nil {
            return nil, err
        }
        f, err := os.Create(name)
        if err != nil {
            closeCallback(fd)
            return nil, err
        }
        addSwiftFile(f, fd, name)
        return f, nil
    }
    return os.Create(name)
}

// Close intercepts Close for Swift operations.
func Close(f *os.File) error {
    if closeCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            err := closeCallback(wf.swiftFD)
            removeSwiftFile(f)
            if err != nil {
                return err
            }
        }
    }
    return f.Close()
}

// Read intercepts Read for Swift operations.
func Read(f *os.File, data []byte) (int, error) {
    if readCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            b, err := readCallback(wf.swiftFD, len(data))
            if err != nil {
                return 0, err
            }
            copy(data, b)
            return len(b), nil
        }
    }
    return f.Read(data)
}

// Write intercepts Write for Swift operations.
func Write(f *os.File, data []byte) (int, error) {
    if writeCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            return writeCallback(wf.swiftFD, data)
        }
    }
    return f.Write(data)
}

// WriteAt intercepts WriteAt for Swift operations.
func WriteAt(f *os.File, data []byte, offset int64) (int, error) {
    if writeAtCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            return writeAtCallback(wf.swiftFD, data, offset)
        }
    }
    return f.WriteAt(data, offset)
}

// Seek intercepts Seek for Swift operations.
func Seek(f *os.File, offset int64, whence int) (int64, error) {
    if seekCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            return seekCallback(wf.swiftFD, offset, whence)
        }
    }
    return f.Seek(offset, whence)
}

// Sync intercepts Sync for Swift operations.
func Sync(f *os.File) error {
    if syncCallback != nil {
        if wf, ok := getSwiftFile(f); ok {
            return syncCallback(wf.swiftFD)
        }
    }
    return f.Sync()
}

// Remove removes the named file or directory.
func Remove(name string) error {
    if removeCallback != nil {
        return removeCallback(name)
    }
    return os.Remove(name)
}

// Mkdir creates a directory named path with the specified permissions.
func Mkdir(name string, perm os.FileMode) error {
    if mkdirCallback != nil {
        return mkdirCallback(name, uint32(perm))
    }
    return os.Mkdir(name, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents.
func MkdirAll(name string, perm os.FileMode) error {
    if mkdirAllCallback != nil {
        return mkdirAllCallback(name, uint32(perm))
    }
    return os.MkdirAll(name, perm)
}

// Stat returns information about the named file or directory.
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

// Chmod changes the mode of the named file to mode.
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

// ReadDir reads the named directory and returns a list of directory entries.
func ReadDir(name string) ([]os.DirEntry, error) {
    if readDirCallback != nil {
        entries, err := readDirCallback(name)
        if err != nil {
            return nil, err
        }
        result := make([]os.DirEntry, len(entries))
        for i, entry := range entries {
            result[i] = &dirEntry{
                name:  entry.Name,
                isDir: entry.IsDir,
            }
        }
        return result, nil
    }
    return os.ReadDir(name)
}

// CreateTemp creates a temporary file in the specified directory.
func CreateTemp(dir, pattern string) (*os.File, error) {
    if createTempCallback != nil {
        result, err := createTempCallback(dir, pattern)
        if err != nil {
            return nil, err
        }
        f, err := os.OpenFile(result.Filename, os.O_RDWR, 0600)
        if err != nil {
            closeCallback(result.Fd)
            return nil, err
        }
        addSwiftFile(f, result.Fd, result.Filename)
        return f, nil
    }
    return os.CreateTemp(dir, pattern)
}

// RemoveAll removes path and any children it contains.
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

// MkdirTemp creates a temporary directory in the specified directory.
func MkdirTemp(dir, pattern string) (string, error) {
    if mkdirTempCallback != nil {
        return mkdirTempCallback(dir, pattern)
    }
    return os.MkdirTemp(dir, pattern)
}

// fileInfo implements os.FileInfo for Stat results.
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

// dirEntry implements os.DirEntry for ReadDir results.
type dirEntry struct {
    name  string
    isDir bool
}

func (de *dirEntry) Name() string              { return de.name }
func (de *dirEntry) IsDir() bool               { return de.isDir }
func (de *dirEntry) Type() os.FileMode         { if de.isDir { return os.ModeDir } else { return 0 } }
func (de *dirEntry) Info() (os.FileInfo, error) { return Stat(de.name) }
