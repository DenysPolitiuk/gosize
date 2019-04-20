package fileentry

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const (
	minFolders     int    = 25
	maxFolders     int    = 50
	minFiles       int    = 1
	maxFiles       int    = 20
	minFileSizeItr int    = 1
	maxFileSizeItr int    = 1000
	minInnerFolder int    = 1
	maxInnerFolder int    = 5
	fileContent    string = "0123456789"
	testFolderName string = "test_target"
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
		for i := 0; i < randomNumber(minFiles, maxFiles); i++ {
			aFile, err := testFileSetup(newFolder)
			if err != nil {
				os.RemoveAll(dir)
				return []string{}, err
			}
			allFiles = append(allFiles, aFile)
		}
		for i := 0; i < randomNumber(minInnerFolder, maxInnerFolder); i++ {
			innerFolderName := strconv.Itoa(int(time.Now().UnixNano()))
			innerFolder := filepath.Join(newFolder, innerFolderName)
			err := os.Mkdir(innerFolder, os.ModePerm)
			if err != nil {
				os.RemoveAll(dir)
				return []string{}, err
			}
			for i := 0; i < randomNumber(minFiles, maxFiles); i++ {
				aFile, err := testFileSetup(innerFolder)
				if err != nil {
					os.RemoveAll(dir)
					return []string{}, err
				}
				allFiles = append(allFiles, aFile)
			}
		}
	}
	return allFiles, nil
}

func testFileSetup(dir string) (string, error) {
	unixTime := time.Now().UnixNano()
	fileName := strconv.FormatInt(unixTime, 10)
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

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	allFiles, err = testFolderSetup(testFolderName)
	if err != nil {
		// if error try to remove existing test_target and try again
		os.RemoveAll(testFolderName)
		allFiles, err = testFolderSetup(testFolderName)
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
		t.Error("Unable to getFolderSize:", err)
		return
	}
	fe, err := NewFileEntry(testFolderName)
	if err != nil {
		t.Error("Unable to NewFileEntry:", err)
		return
	}
	err = fe.FillContent()
	if err != nil {
		t.Error("Error during FillContent:", err)
		return
	}
	if float64(fe.Size) != float64(totalSize) {
		t.Errorf("Total size for NewFileEntry is %v but it is %v for test result", float64(fe.Size), totalSize)
	}
	files, err := ioutil.ReadDir(testFolderName)
	if err != nil {
		t.Error("Unable to ReadDir:", err)
		return
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		fullName := filepath.Join(testFolderName, f.Name())
		fSize, err := getFolderSize(fullName)
		if err != nil {
			t.Error("Unable to getFolderSize:", err)
			continue
		}
		if f, ok := fe.Content[f.Name()]; !ok {
			t.Errorf("Unable to find %v in FileEntry Content", fullName)
		} else {
			if float64(fSize) != float64(f.Size) {
				t.Errorf("FieEntry Size for %v is %v instead of %v", fullName, f.Size, fSize)
			}
		}
	}
}
