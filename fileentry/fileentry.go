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
	Msg           string
	MethodName    string
	InternalError error
	Severity      ErrorSeverity
}

type ErrorSeverity string

const (
	Critical ErrorSeverity = "critical"
	Normal   ErrorSeverity = "normal"
)

func (err *BasicError) Error() string {
	if err.InternalError != nil {
		return fmt.Sprintf("%v at %v", err.InternalError, err.MethodName)
	}
	return fmt.Sprintf("%v at %v", err.Msg, err.MethodName)
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
		return &FileEntry{}, &BasicError{InternalError: err, MethodName: "FileEntry.NewFileEntry os.Stat()", Severity: Critical}
	}
	if !fi.IsDir() {
		return &FileEntry{}, &BasicError{Msg: fmt.Sprintf("%s: is not a directory", root),
			MethodName: "FileEntry.NewFileEntry IsDir()", InternalError: nil, Severity: Critical}
	}
	fe := &FileEntry{Type: Directory, Parent: &FileEntry{}, Name: root, FullPath: root, Content: make(map[string]*FileEntry)}
	return fe, nil
}

func (fe *FileEntry) FillContent() error {
	if fe.Type != Directory {
		return &BasicError{Msg: fmt.Sprintf("%s is %v and not %v", fe.Name, fe.Type, Directory),
			MethodName: "FileEntry.FillContent", InternalError: nil, Severity: Normal}
	}
	files, err := ioutil.ReadDir(fe.FullPath)
	if err != nil {
		return &BasicError{InternalError: err, MethodName: "FileEntry.FillContent ioutil.ReadDir()", Severity: Normal}
	}
	for _, f := range files {
		fullPath := filepath.Join(fe.FullPath, f.Name())
		switch f.IsDir() {
		case true:
			// need to go deeper and process inner directory
			dir := &FileEntry{Type: Directory, Parent: fe, Name: f.Name(), FullPath: fullPath, Content: make(map[string]*FileEntry)}
			fe.Content[f.Name()] = dir
			err = dir.FillContent()
			if err != nil {
				if e, ok := err.(*BasicError); ok {
					if e.Severity == Critical {
						return e
					} else {
						fmt.Println(e)
					}
				} else {
					return err
				}
			} else {
				fe.Size += dir.Size
			}
		case false:
			size := ByteSize(float64(f.Size()))
			fe.Content[f.Name()] = &FileEntry{Type: File, Parent: fe, Name: f.Name(), FullPath: fullPath, Size: size}
			fe.Size += size
		}
	}
	return nil
}

func (fe *FileEntry) Search(name string, t FileType) ([]*FileEntry, error) {
	result := make([]*FileEntry, 0, 5)
	if fe.Type == t || t == Unknown {
		if fe.Name == name {
			result = append(result, fe)
		}
	}
	if fe.Type != Directory {
		return result, nil
	}
	// continue to process content inside a directory
	for _, c := range fe.Content {
		r, err := c.Search(name, t)
		if err != nil {
			if e, ok := err.(*BasicError); ok {
				if e.Severity != Critical {
					fmt.Println(e)
				} else {
					return result, e
				}
			} else {
				return result, err
			}
		}
		result = append(result, r...)
	}
	return result, nil
}

// TODO: Implement a method to convert FileEntry.Content into a slice with all FileEntries from the Content
func (fe *FileEntry) Flatten() ([]*FileEntry, error) {
	return []*FileEntry{}, nil
}
