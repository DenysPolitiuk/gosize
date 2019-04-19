package fileentry

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ByteSize float64

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

type BasicError struct {
	Msg string
}

func (err *BasicError) Error() string {
	return err.Msg
}

type FileType string

const (
	Unknown   FileType = "unknown"
	File      FileType = "file"
	Directory FileType = "directory"
)

type FileEntry struct {
	Type     FileType
	Parent   *FileEntry
	Name     string
	FullPath string
	Content  map[string]*FileEntry
	Size     ByteSize
}

func NewFileEntry(root string) (*FileEntry, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return &FileEntry{}, err
	}
	if !fi.IsDir() {
		return &FileEntry{}, &BasicError{fmt.Sprintf("%s: is not a directory", root)}
	}
	fe := &FileEntry{Type: Directory, Parent: &FileEntry{}, Name: root, FullPath: root, Content: make(map[string]*FileEntry)}
	return fe, nil
}

func (fe *FileEntry) FillContent() error {
	if fe.Type != Directory {
		return &BasicError{fmt.Sprintf("%s is %v and not %v", fe.Name, fe.Type, Directory)}
	}
	files, err := ioutil.ReadDir(fe.FullPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		fullPath := filepath.Join(fe.Name, f.Name())
		switch f.IsDir() {
		case true:
			// need to go deeper and process inner directory
			dir := &FileEntry{Type: Directory, Parent: fe, Name: f.Name(), FullPath: fullPath, Content: make(map[string]*FileEntry)}
			fe.Content[f.Name()] = dir
			err = dir.FillContent()
			if err != nil {
				return err
			}
			fe.Size += dir.Size
		case false:
			size := ByteSize(float64(f.Size()))
			fe.Content[f.Name()] = &FileEntry{Type: File, Parent: fe, Name: f.Name(), FullPath: fullPath, Size: size}
			fe.Size += size
		}
	}
	return nil
}
