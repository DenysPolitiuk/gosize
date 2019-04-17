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
import _ "io/ioutil"

// TODO: use for going through folder inner files

const (
	minFolders     int    = 25
	maxFolders     int    = 50
	minFiles       int    = 1
	maxFiles       int    = 20
	minFileSizeItr int    = 1
	maxFileSizeItr int    = 1000
	fileContent    string = "0123456789"
)

func testFolderSetup(dir string) ([]string, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return []string{}, err
	}
	allFiles := make([]string, 1, maxFolders*maxFiles)
	f, err := testFileSetup(dir)
	if err != nil {
		os.RemoveAll(dir)
		return []string{}, err
	}
	// set up and add one file for fail tests
	allFiles[0] = filepath.Join(dir, f)
	for i := 0; i < randomNumber(minFolders, maxFolders); i++ {
		folderName := strconv.Itoa(int(time.Now().UnixNano()))
		newFolder := filepath.Join(dir, folderName)
		err = os.Mkdir(newFolder, os.ModePerm)
		if err != nil {
			os.RemoveAll(dir)
			return []string{}, err
		}
		for j := 0; j < randomNumber(minFiles, maxFiles); j++ {
			aFile, err := testFileSetup(newFolder)
			if err != nil {
				os.RemoveAll(dir)
				return []string{}, err
			}
			allFiles = append(allFiles, aFile)
		}
	}
	return allFiles, nil
}

func testFileSetup(dir string) (string, error) {
	unixTime := time.Now().UnixNano()
	fileName := strconv.FormatInt(unixTime, 10) //fmt.Sprintf("%v", unixTime)
	f, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return "", err
	}
	defer f.Close()
	for i := 0; i < randomNumber(minFileSizeItr, maxFileSizeItr); i++ {
		_, err := f.WriteString(fileContent)
		if err != nil {
			return "", err
		}
	}
	return fileName, nil
}

func randomNumber(min int, max int) int {
	return rand.Intn(max-min) + min
}

func getFolderSize(folder string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(folder, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return err
	})
	return totalSize, err
}

var allFiles []string

const testFolderName = "test_target"

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	allFiles, err = testFolderSetup(testFolderName)
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
				t.Errorf("NewFileEntry(%v) returned nil instead of %v", c.in, got)
			}
		}
	}
}

func TestFillContent(t *testing.T) {
	totalSize, err := getFolderSize(testFolderName)
	if err != nil {
		t.Error("Unable to getFolderSize", err)
		return
	}
	fe, err := NewFileEntry(testFolderName)
	if err != nil {
		t.Error("Unable to NewFileEntry", err)
		return
	}
	err = fe.FillContent()
	if err != nil {
		t.Error("Error during FillContent", err)
		return
	}
	if float64(fe.Size) != float64(totalSize) {
		t.Errorf("Total size for NewFileEntry is %v but it is %v for test result", float64(fe.Size), totalSize)
	}
	// for future tests
	// files, err := ioutil.ReadDir(testFolderName)
	// if err != nil {
	// t.Error("Unable to ReadDir", err)
	// return
	// }
	// for _, f := range files {
	// if !f.IsDir() {
	// continue
	// }
	// fSize, err := getFolderSize(f.Name())
	// if err != nil {
	// t.Error("Unable to getFolderSize", err)
	// return
	// }
	// }
}
