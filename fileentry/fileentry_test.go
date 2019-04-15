package fileentry

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func testFolderSetup(dir string) ([]string, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return []string{}, err
	}
	allFiles := make([]string, 1, 50)
	f, err := testFileSetup(dir)
	if err != nil {
		os.RemoveAll(dir)
		return []string{}, err
	}
	// set up and add one file for fail tests
	allFiles[0] = filepath.Join(dir, f)
	for i := 0; i < randomNumber(1, 20); i++ {
		folderName := strconv.Itoa(int(time.Now().UnixNano()))
		newFolder := filepath.Join(dir, folderName)
		err = os.Mkdir(newFolder, os.ModePerm)
		if err != nil {
			os.RemoveAll(dir)
			return []string{}, err
		}
		for j := 0; j < randomNumber(1, 20); j++ {
			_, err = testFileSetup(newFolder)
			if err != nil {
				os.RemoveAll(dir)
				return []string{}, err
			}
		}
	}
	return allFiles, nil
}

func testFileSetup(dir string) (string, error) {
	unixTime := time.Now().UnixNano()
	fileName := fmt.Sprintf("%v.txt", unixTime)
	f, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return "", err
	}
	defer f.Close()
	for i := 0; i < randomNumber(1, 100); i++ {
		_, err := f.WriteString("0123456789")
		if err != nil {
			return "", err
		}
	}
	return fileName, nil
}

func randomNumber(min int, max int) int {
	return rand.Intn(max-min) + min
}

var allFiles []string

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	allFiles, err = testFolderSetup("test_target")
	if err != nil {
		// if error try to remove existing test_target and try again
		os.RemoveAll("test_target")
		allFiles, err = testFolderSetup("test_target")
		if err != nil {
			fmt.Println("Error in init\n", err)
			os.Exit(1)
		}
	}
}

func TestNewFileEntry(t *testing.T) {
	cases := []struct {
		in    string
		out   *FileEntry
		valid bool
	}{
		{"nonexistingfolder", nil, false},
		{allFiles[0], nil, false},
	}
	for _, c := range cases {
		got, err := NewFileEntry(c.in)
		switch c.valid {
		case true:
		case false:
			if err == nil {
				t.Errorf("Parser(%v) returned nil instead of %v", c.in, got)
			}
		}
	}
}

func TestFillContent(t *testing.T) {
	cases := []struct {
		in    string
		out   *FileEntry
		valid bool
	}{}
	for _, _ = range cases {
	}
	t.Errorf("Not implemented")
}
