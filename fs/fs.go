package fs

import (
	"os"
)

type FS interface {
	Open(string) (File, error)
	OpenFile(string, int, os.FileMode) (File, error)
	Stat(string) (os.FileInfo, error)
}

type File interface {
	Close() error
	Read([]byte) (int, error)
	WriteString(string) (int, error)
}

type osFS struct{}

func FromOSFS() FS {
	return &osFS{}
}

func (osfs *osFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (osfs *osFS) Open(path string) (File, error) {
	return os.Open(path)
}

func (osfs *osFS) OpenFile(path string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(path, flag, perm)
}
