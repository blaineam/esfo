package esfo

import (
    "os"
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

// File represents a file handle for esfo operations.
type File struct {
    fd   int64  // File descriptor ID for Swift
    name string // File name for compatibility
}

// Fd returns the file descriptor ID.
func (f *File) Fd() uintptr {
    return uintptr(f.fd)
}

// Name returns the file name.
func (f *File) Name() string {
    return f.name
}

// Write writes data to the file descriptor.
func (f *File) Write(data []byte) (int, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.Write(f.fd, data)
    }
    f2, err := os.OpenFile(f.name, os.O_WRONLY, 0)
    if err != nil {
        return 0, err
    }
    defer f2.Close()
    return f2.Write(data)
}

// WriteString writes a string to the file descriptor.
func (f *File) WriteString(s string) (int, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.Write(f.fd, []byte(s))
    }
    f2, err := os.OpenFile(f.name, os.O_WRONLY, 0)
    if err != nil {
        return 0, err
    }
    defer f2.Close()
    return f2.WriteString(s)
}

// Close closes the file descriptor.
func (f *File) Close() error {
    if fileSystemHandler != nil {
        return fileSystemHandler.Close(f.fd)
    }
    f2, err := os.Open(f.name)
    if err != nil {
        return err
    }
    return f2.Close()
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
func OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
    if fileSystemHandler != nil {
        fd, err := fileSystemHandler.OpenFile(name, flag, uint32(perm))
        if err != nil {
            return nil, err
        }
        return &File{fd: fd, name: name}, nil
    }
    f, err := os.OpenFile(name, flag, perm)
    if err != nil {
        return nil, err
    }
    return &File{fd: int64(f.Fd()), name: name}, nil
}

// Create creates or truncates the named file.
func Create(name string) (*File, error) {
    if fileSystemHandler != nil {
        fd, err := fileSystemHandler.Create(name)
        if err != nil {
            return nil, err
        }
        return &File{fd: fd, name: name}, nil
    }
    f, err := os.Create(name)
    if err != nil {
        return nil, err
    }
    return &File{fd: int64(f.Fd()), name: name}, nil
}

// Read reads up to count bytes from the file descriptor.
func Read(f *File, count int) ([]byte, error) {
    if fileSystemHandler != nil {
        return fileSystemHandler.Read(f.fd, count)
    }
    f2, err := os.Open(f.name)
    if err != nil {
        return nil, err
    }
    defer f2.Close()
    b := make([]byte, count)
    n, err := f2.Read(b)
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

// CreateTemp creates a temporary file in the specified directory with the given pattern.
func CreateTemp(dir, pattern string) (*File, error) {
    if fileSystemHandler != nil {
        filename, fd, err := fileSystemHandler.CreateTemp(dir, pattern)
        if err != nil {
            return nil, err
        }
        return &File{fd: fd, name: filename}, nil
    }
    f, err := os.CreateTemp(dir, pattern)
    if err != nil {
        return nil, err
    }
    return &File{fd: int64(f.Fd()), name: f.Name()}, nil
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

// MkdirTemp creates a temporary directory in the specified directory with the given pattern.
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