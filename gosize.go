package main

import (
	"fmt"
	"github.com/denyspolitiuk/clparser"
	"github.com/denyspolitiuk/gosize/fileentry"
	"os"
)

func printHelp() {
	// TODO: Add help print
	fmt.Println("TODO")
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
	for k, v := range mpArgs {
		switch k {
		case "-t", "--target":
			rootName = v
		case "-h", "--help":
			printHelp()
			return
		}
	}
	mp, err := fileentry.NewFileEntry(rootName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = mp.FillContent()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mp.Size)
}
