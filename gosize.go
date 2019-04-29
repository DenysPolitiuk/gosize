package main

import (
	"bufio"
	"fmt"
	"github.com/denyspolitiuk/clparser"
	"github.com/denyspolitiuk/gosize/fileentry"
	"os"
	"strings"
)

const (
	interactiveEntryPrintLimit = 20
)

func printHelp() {
	fmt.Println("Application to help with file/folder size analysis")
	fmt.Println("Commands :")
	fmt.Println("\t-t NAME\t--target=NAME\t Target folder to server as root, required")
	fmt.Println("\t-h\t--help\t\t Print help")
	fmt.Println("\t-s NAME\t--search=NAME\tSearch for one specific file/folder and print it's size")
	fmt.Println("\t-i\t\t\tInteractive view")
	fmt.Println("\t-e NAME\t--encode=NAME\tSave target to a file")
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		printHelp()
		return
	}
	mpArgs, err := clparser.Parse(args)
	if err != nil {
		fmt.Println(err)
		return
	}
	var rootName string
	var search string
	inter := false
	var encodePath string
	for k, v := range mpArgs {
		switch k {
		case "-t", "--target":
			rootName = v
		case "-h", "--help":
			printHelp()
			return
		case "-s", "--search":
			search = v
		case "-i":
			inter = true
		case "-e", "--encode":
			encodePath = v
		}
	}
	fe, err := fileentry.NewFileEntry(rootName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = fe.FillContent()
	if err != nil {
		fmt.Println(err)
		return
	}
	// if doing interactive only do it and leave the program
	if inter {
		err := interactive(fe)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	if rootName == "" {
		fmt.Println("root directory is required")
		printHelp()
		return
	}
	if encodePath != "" {
		err := fileentry.Save(encodePath, fe)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	if search != "" {
		result, err := fe.Search(search, fileentry.Unknown)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, r := range result {
			fmt.Printf("%v has size %v\n", r.Name, r.Size)
		}
		return
	}
	fmt.Println(fe.Size)
}

func printInteractiveMsg() {
	fmt.Println("Use wasd to navigate through folders/files")
	fmt.Println("Use t to load new folder, o to switch sort method")
	fmt.Println("Use q for quit, h to see this message again")
}

func prettyPrintFileEntry(root *fileentry.FileEntry, sortedContent []*fileentry.FileEntry, offset int, targetIndex int) bool {
	fmt.Printf("%s | %s\n", root.Name, root.Size)
	printedCounter := 0
	skipped := false
	printedAll := false
	if offset != 0 {
		fmt.Println("...")
	}
	for i, f := range sortedContent {
		if i < offset {
			continue
		}
		if printedCounter > interactiveEntryPrintLimit {
			skipped = true
			break
		}
		if i == targetIndex {
			fmt.Print("->")
		}
		fmt.Printf("%10s | %20s | %s\n", f.Type, f.Name, f.Size)
		printedCounter++
	}
	if skipped {
		fmt.Println("...")
	} else {
		printedAll = true
	}
	return printedAll
}

func interactive(fileEntry *fileentry.FileEntry) error {
	reader := bufio.NewReader(os.Stdin)
	exit := false
	offset := 0
	targetIndex := 0
	sortType := fileentry.Name
	sortedContent := fileEntry.GetSortedContent(sortType)
	printInteractiveMsg()
	for {
		printedAll := prettyPrintFileEntry(fileEntry, sortedContent, offset, targetIndex)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		userInput = strings.Trim(userInput, "\n")
		switch userInput {
		case "w":
			if targetIndex > 0 {
				targetIndex--
			}
			if offset > 0 {
				offset--
			}
		case "a":
			parentEntry := fileEntry.Parent
			if parentEntry == nil || parentEntry.Name == "" {
				continue
			}
			fileEntry = parentEntry
			sortedContent = fileEntry.GetSortedContent(sortType)
			targetIndex = 0
			offset = 0
		case "s":
			if targetIndex < (len(sortedContent) - 1) {
				targetIndex++
			}
			if !printedAll {
				offset++
			}
		case "d":
			targetEntry := sortedContent[targetIndex]
			if targetEntry.Type != fileentry.Directory {
				continue
			}
			fileEntry = targetEntry
			sortedContent = fileEntry.GetSortedContent(sortType)
			targetIndex = 0
			offset = 0
		case "t":
			fmt.Println("Enter new target :")
			userInput, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			userInput = strings.Trim(userInput, "\n")
			fmt.Printf("New target is : %s\n", userInput)
			fileEntry, err = fileentry.NewFileEntry(userInput)
			if err != nil {
				return err
			}
			err = fileEntry.FillContent()
			if err != nil {
				return err
			}
			sortedContent = fileEntry.GetSortedContent(sortType)
			targetIndex = 0
			offset = 0
		case "o":
			fmt.Println("Sort by (n)ame or (s)ize ?")
			userInput, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			userInput = strings.Trim(userInput, "\n")
			oldSortType := sortType
			switch userInput {
			case "n", "name":
				sortType = fileentry.Name
			case "s", "size":
				sortType = fileentry.Size
			default:
				continue
			}
			if oldSortType == sortType {
				continue
			}
			sortedContent = fileEntry.GetSortedContent(sortType)
		case "h":
			printInteractiveMsg()
		case "q":
			exit = true
		}
		if exit {
			break
		}
	}
	return nil
}
