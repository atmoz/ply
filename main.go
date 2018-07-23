package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
)

var site Site

func main() {
	usage := `ply - recursive markdown to HTML converter

Usage: ply [clean] [<path>] [-h|--help]`

	var err error
	args, _ := docopt.ParseDoc(usage)
	argClean, _ := args.Bool("clean")
	argPath, _ := args.String("<path>")

	if argPath == "" {
		argPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	} else {
		argPath = filepath.Clean(argPath)
	}

	site.Path = argPath
	site.Init()

	if argClean {
		err = site.Clean()
	} else {
		err = site.Build()
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
