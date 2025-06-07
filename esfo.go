package esfo

import (
    "os"
    "sync"
    "time"
)

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
    swiftFD int64    // Swift file descriptor ID
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

// Write intercepts Write for Swift operations.
func Write(f *os.File, data []byte) (int, error) {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return fileSystemHandler.Write(wf.swiftFD, data)
        }
    }
    return f.Write(data)
}

// WriteString intercepts WriteString for Swift operations.
func WriteString(f *os.File, s string) (int, error) {
    return Write(f, []byte(s))
}

// WriteAt intercepts WriteAt for Swift operations.
func WriteAt(f *os.File, data []byte, offset int64) (int, error) {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return fileSystemHandler.WriteAt(wf.swiftFD, data, offset)
        }
    }
    return f.WriteAt(data, offset)
}

// Read intercepts Read for Swift operations.
func Read(f *os.File, data []byte) (int, error) {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            b, err := fileSystemHandler.Read(wf.swiftFD, len(data))
            if err != nil {
                return 0, err
            }
            copy(data, b)
            return len(b), nil
        }
    }
    return f.Read(data)
}

// Close intercepts Close for Swift operations.
func Close(f *os.File) error {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            err := fileSystemHandler.Close(wf.swiftFD)
            removeSwiftFile(f)
            if err != nil {
                return err
            }
        }
    }
    return f.Close()
}

// Seek intercepts Seek for Swift operations.
func Seek(f *os.File, offset int64, whence int) (int64, error) {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return fileSystemHandler.Seek(wf.swiftFD, offset, whence)
        }
    }
    return f.Seek(offset, whence)
}

// Sync intercepts Sync for Swift operations.
func Sync(f *os.File) error {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return fileSystemHandler.Sync(wf.swiftFD)
        }
    }
    return f.Sync()
}

// Name intercepts Name for Swift operations.
func Name(f *os.File) string {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return wf.name
        }
    }
    return f.Name()
}

// FileSystemHandler is the interface implemented by Swift for file operations.
type FileSystemHandler interface {
    WriteFile(filename string, data []byte, perm uint32) error
    ReadFile(filename string) ([]byte, error)
    OpenFile(name string, flag int, perm uint32) (int64, error) // Returns file descriptor ID
    Create(name string) (int64, error)                          // Returns file descriptor ID
    Close(fd int64) error
    Read(fd int64, count int) ([]byte, error)
    Write(fd int64, data []byte) (int, error)
    WriteAt(fd int64, data []byte, offset int64) (int, error)
    Seek(fd int64, offset int64, whence int) (int64, error)
    Sync(fd int64) error
    Remove(name string) error
    Mkdir(name string, perm uint32) error
    MkdirAll(name string, perm uint32) error
    Stat(name string) (FileInfo, error)
    Chmod(name string, mode uint32) error
    Rename(oldpath, newpath string) error
    ReadDir(name string) ([]DirEntry, error)
    CreateTemp(dir, pattern string) (string, int64, error) // Returns filename and file descriptor ID
    RemoveAll(path string) error
    ReadLink(name string) (string, error)
    MkdirTemp(dir, pattern string) (string, error) // Returns directory name
}

var fileSystemHandler FileSystemHandler

// SetFileSystemHandler sets the Swift implementation for file operations.
func SetFileSystemHandler(handler FileSystemHandler) {
    fileSystemHandler = handler
}

// WriteFile writes data to the named file with the given permissions.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.WriteFile(filename, data, uint32(perm))
    }
    return os.WriteFile(filename, data, perm)
}

// ReadFile reads the named file and returns its contents.
func ReadFile(filename string) ([]byte, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.ReadFile(filename)
    }
    return os.ReadFile(filename)
}

// OpenFile opens the named file with specified flag and permissions.
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
    if fileSystemHandler != nil {
        fd, err := fileSystemHandler.OpenFile(name, flag, uint32(perm))
        if err != nil {
            return nil, err
        }
        f, err := os.OpenFile(name, flag, perm)
        if err != nil {
            fileSystemHandler.Close(fd)
            return nil, err
        }
        addSwiftFile(f, fd, name)
        return f, nil
    }
    return os.OpenFile(name, flag, perm)
}

// Create creates or truncates the named file.
func Create(name string) (*os.File, error) {
    if fileSystemHandler != nil {
        fd, err := fileSystemHandler.Create(name)
        if err != nil {
            return nil, err
        }
        f, err := os.Create(name)
        if err != nil {
            fileSystemHandler.Close(fd)
            return nil, err
        }
        addSwiftFile(f, fd, name)
        return f, nil
    }
    return os.Create(name)
}

// CreateTemp creates a temporary file in the specified directory.
func CreateTemp(dir, pattern string) (*os.File, error) {
    if fileSystemHandler != nil {
        filename, fd, err := fileSystemHandler.CreateTemp(dir, pattern)
        if err != nil {
            return nil, err
        }
        f, err := os.OpenFile(filename, os.O_RDWR, 0600)
        if err != nil {
            fileSystemHandler.Close(fd)
            return nil, err
        }
        addSwiftFile(f, fd, filename)
        return f, nil
    }
    return os.CreateTemp(dir, pattern)
}

// Read reads up to count bytes from the file.
func ReadFileData(f *os.File, count int) ([]byte, error) {
    if fileSystemHandler != nil {
        if wf, ok := getSwiftFile(f); ok {
            return fileSystemHandler.Read(wf.swiftFD, count)
        }
    }
    b := make([]byte, count)
    n, err := f.Read(b)
    return b[:n], err
}

// Remove removes the named file or directory.
func Remove(name string) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.Remove(name)
    }
    return os.Remove(name)
}

// Mkdir creates a directory named path with the specified permissions.
func Mkdir(name string, perm os.FileMode) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.Mkdir(name, uint32(perm))
    }
    return os.Mkdir(name, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents.
func MkdirAll(name string, perm os.FileMode) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.MkdirAll(name, uint32(perm))
    }
    return os.MkdirAll(name, perm)
}

// Stat returns information about the named file or directory.
func Stat(name string) (os.FileInfo, error) {
    if fileSystemHandler != nil {
        fi, err := fileSystemHandler.Stat(name)
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

// Chmod changes the mode of the named file.
func Chmod(name string, mode os.FileMode) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.Chmod(name, uint32(mode))
    }
    return os.Chmod(name, mode)
}

// Rename renames (moves) oldpath to newpath.
func Rename(oldpath, newpath string) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.Rename(oldpath, newpath)
    }
    return os.Rename(oldpath, newpath)
}

// ReadDir reads the named directory and returns a list of directory entries.
func ReadDir(name string) ([]os.DirEntry, error) {
    if fileSystemHandler != nil {
        entries, err := fileSystemHandler.ReadDir(name)
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

// RemoveAll removes path and any children it contains.
func RemoveAll(path string) error {
    if fileSystemHandler != nil {
        return fileSystemHandler.RemoveAll(path)
    }
    return os.RemoveAll(path)
}

// ReadLink returns the destination of the named symbolic link.
func ReadLink(name string) (string, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.ReadLink(name)
    }
    return os.Readlink(name)
}

// MkdirTemp creates a temporary directory in the specified directory.
func MkdirTemp(dir, pattern string) (string, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.MkdirTemp(dir, pattern)
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

func (de *dirEntry) Name() string          { return de.name }
func (de *dirEntry) IsDir() bool           { return de.isDir }
func (de *dirEntry) Type() os.FileMode     { if de.isDir { return os.ModeDir } else { return 0 } }
func (de *dirEntry) Info() (os.FileInfo, error) {
    return Stat(de.name)
}