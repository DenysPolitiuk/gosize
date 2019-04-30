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
	minFolders     int    = 2
	maxFolders     int    = 5
	minFiles       int    = 2
	maxFiles       int    = 5
	minFileSizeItr int    = 1
	maxFileSizeItr int    = 5
	minFolderDepth int    = 2
	maxFolderDepth int    = 3
	fileContent    string = "0123456789"
	testFolderName string = "test_target"
	encodingName   string = "test.gob"
)

var (
	allFiles   []string
	allFolders []string
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
	allFiles[0] = f
	innerFiles, innerFolders, err := createDeepFolders(dir, randomNumber(minFolderDepth, maxFolderDepth))
	if err != nil {
		os.RemoveAll(dir)
		return []string{}, err
	}
	allFiles = append(allFiles, innerFiles...)
	allFolders = append(allFolders, innerFolders...)
	return allFiles, nil
}

func createDeepFolders(startingFolder string, depth int) ([]string, []string, error) {
	if depth < 1 {
		return []string{}, []string{}, nil
	}
	returnedFiles := make([]string, 0, 10)
	returnedFolders := make([]string, 0, 10)
	depth--
	for i := 0; i < randomNumber(minFolders, maxFolders); i++ {
		folderName := strconv.Itoa(int(time.Now().UnixNano()))
		newFolder := filepath.Join(startingFolder, folderName)
		err := os.Mkdir(newFolder, os.ModePerm)
		if err != nil {
			return []string{}, []string{}, err
		}
		returnedFolders = append(returnedFolders, folderName)
		files, err := createRandomNumberOfFiles(newFolder)
		if err != nil {
			return []string{}, []string{}, err
		}
		returnedFiles = append(returnedFiles, files...)
		innerFiles, innerFolders, err := createDeepFolders(newFolder, depth)
		if err != nil {
			return returnedFiles, returnedFolders, err
		}
		returnedFiles = append(returnedFiles, innerFiles...)
		returnedFolders = append(returnedFolders, innerFolders...)
	}
	return returnedFiles, returnedFolders, nil
}

func createRandomNumberOfFiles(folderName string) ([]string, error) {
	var allFiles []string
	for i := 0; i < randomNumber(minFiles, maxFiles); i++ {
		aFile, err := testFileSetup(folderName)
		if err != nil {
			return []string{}, err
		}
		allFiles = append(allFiles, aFile)
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

func TestSearch(t *testing.T) {
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
	for _, f := range allFiles {
		uResult, err := fe.Search(f, Unknown)
		if err != nil {
			t.Error(err)
			continue
		}
		fResult, err := fe.Search(f, File)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(uResult) != 1 {
			t.Errorf("Got %v entries instead of %v for FileEntry.Search() for file %v", len(uResult), 1, f)
			continue
		}
		if uResult[0].Name != f {
			t.Errorf("Got %v name for FileEntry.Search() instead of %v", uResult[0].Name, f)
			continue
		}
		if len(fResult) != 1 {
			t.Errorf("Got %v entries instead of %v for FileEntry.Search() for file %v", len(fResult), 1, f)
			continue
		}
		if fResult[0].Name != f {
			t.Errorf("Got %v name for FileEntry.Search() instead of %v", fResult[0].Name, f)
			continue
		}
	}
	for _, f := range allFolders {
		uResult, err := fe.Search(f, Unknown)
		if err != nil {
			t.Error(err)
			continue
		}
		fResult, err := fe.Search(f, Directory)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(uResult) != 1 {
			t.Errorf("Got %v entries instead of %v for FileEntry.Search() for directory %v", len(uResult), 1, f)
			continue
		}
		if uResult[0].Name != f {
			t.Errorf("Got %v name for FileEntry.Search() instead of %v", uResult[0].Name, f)
			continue
		}
		if len(fResult) != 1 {
			t.Errorf("Got %v entries instead of %v for FileEntry.Search() for directory %v", len(fResult), 1, f)
			continue
		}
		if fResult[0].Name != f {
			t.Errorf("Got %v name for FileEntry.Search() instead of %v", fResult[0].Name, f)
			continue
		}
	}
}

func TestFlatten(t *testing.T) {
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
	flat, err := fe.Flatten(Unknown, 0)
	if err != nil {
		t.Error("Error during Flatten Unknown:", err)
	}
	allEntries := make([]string, 0, len(allFiles)+len(allFolders))
	allEntries = append(allEntries, allFiles...)
	allEntries = append(allEntries, allFolders...)
	for _, entry := range allEntries {
		found := false
		for _, fe := range flat {
			if fe.Name == entry {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in Flatten Unknown result", entry)
		}
	}
	flat, err = fe.Flatten(File, 0)
	if err != nil {
		t.Error("Error during Flatten File:", err)
	}
	for _, entry := range allFiles {
		found := false
		for _, fe := range flat {
			if fe.Name == entry {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in Flatten File result", entry)
		}
	}
	for _, entry := range allFolders {
		found := false
		for _, fe := range flat {
			if fe.Name == entry {
				found = true
				break
			}
		}
		if found {
			t.Errorf("Able to find %v in Flatten File result even though it's a folder", entry)
		}
	}
	flat, err = fe.Flatten(Directory, 0)
	if err != nil {
		t.Error("Error during Flatten Directory:", err)
	}
	for _, entry := range allFiles {
		found := false
		for _, fe := range flat {
			if fe.Name == entry {
				found = true
				break
			}
		}
		if found {
			t.Errorf("Able to find %v in Flatten Directory result even though it's a file", entry)
		}
	}
	for _, entry := range allFolders {
		found := false
		for _, fe := range flat {
			if fe.Name == entry {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in Flatten Directory result", entry)
		}
	}
	// basic depth test
	files, err := ioutil.ReadDir(testFolderName)
	if err != nil {
		t.Error("Unable to ReadDir:", err)
		return
	}
	inner, err := fe.Flatten(Unknown, 1)
	if err != nil {
		t.Error("Unable to fe.Flatten Unknown 1", err)
		return
	}
	if len(inner) != len(files) {
		t.Errorf("Number of entry at depth 1 from FileEntry.Flatten Unknown 1 is %v instead of %v", len(inner), len(files))
		return
	}
	for _, i := range inner {
		found := false
		for _, f := range files {
			if i.Name == f.Name() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File Entry %v is not in list of entry at depth 1", i.Name)
		}
	}
}

func TestGetSortedContent(t *testing.T) {
	original := &FileEntry{}
	original.Content = make(map[string]*FileEntry)
	for i := 0; i < 10; i++ {
		innerContent := &FileEntry{}
		innerContent.Name = strconv.Itoa(i)
		innerContent.Size = ByteSize(i)
		original.Content[innerContent.Name] = innerContent
	}
	sortedName := original.GetSortedContent(Name)
	sortedSize := original.GetSortedContent(Size)
	for i, e := range sortedSize {
		size := ByteSize(9 - i)
		if e.Size != size {
			t.Errorf("Should have size of %v but have %v\n", size, e.Size)
		}
	}
	for i, e := range sortedName {
		name := strconv.Itoa(i)
		if e.Name != name {
			t.Errorf("Should have name %v but have %v\n", name, e.Name)
		}
	}
}

func TestOpenSave(t *testing.T) {
	fe, err := NewFileEntry(testFolderName)
	if err != nil {
		t.Errorf("Got error in NewFileEntry %v\n", err)
		return
	}
	err = fe.FillContent()
	if err != nil {
		t.Errorf("Got error in FillContent %v\n", err)
		return
	}
	fe2, err := NewFileEntry(testFolderName)
	if err != nil {
		t.Errorf("Got error in NewFileEntry %v\n", err)
		return
	}
	err = fe2.FillContent()
	if err != nil {
		t.Errorf("Got error in FillContent %v\n", err)
		return
	}
	feFlat, err := fe.Flatten(Unknown, 0)
	if err != nil {
		t.Errorf("Got error in Flatten %v\n", err)
		return
	}
	fe2Flat, err := fe2.Flatten(Unknown, 0)
	if err != nil {
		t.Errorf("Got error in Flatten %v\n", err)
		return
	}
	for _, e := range feFlat {
		found := false
		for _, e2 := range fe2Flat {
			if e.Name == e2.Name && e.Size == e2.Size && e.FullPath == e.FullPath && e.Parent.Name == e2.Parent.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in FileEntry duplicate\n", e.Name)
		}
	}
	err = Save(encodingName, fe2)
	if err != nil {
		t.Errorf("Error in Save %v\n", err)
		return
	}
	// check to see if fe2 have changes after saving
	fe2Flat, err = fe2.Flatten(Unknown, 0)
	if err != nil {
		t.Errorf("Error in Flatten %v\n", err)
		return
	}
	for _, e := range feFlat {
		found := false
		for _, e2 := range fe2Flat {
			if e.Name == e2.Name && e.Size == e2.Size && e.FullPath == e.FullPath && e.Parent.Name == e2.Parent.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in FileEntry duplicate after Save\n", e.Name)
		}
	}
	newFe, err := Open(encodingName)
	newFeFlat, err := newFe.Flatten(Unknown, 0)
	if err != nil {
		t.Errorf("Error in Flatten %v\n", err)
		return
	}
	for _, e := range feFlat {
		found := false
		for _, e2 := range newFeFlat {
			if e.Name == e2.Name && e.Size == e2.Size && e.FullPath == e.FullPath && e.Parent.Name == e2.Parent.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Not able to find %v in FileEntry duplicate after Open\n", e.Name)
		}
	}
}
