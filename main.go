package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/atmoz/ply/fileutil"
	"github.com/docopt/docopt-go"
)

var site Site

func main() {
	usage := `ply - recursive markdown to HTML converter

Usage:
  ply <source-path> <target-path> [options]
  ply -h|--help

Options:
  --include-markdown  Include markdown files in target
  --include-template  Include template files in target
  --ignore=<regex>    File names to ignore (defaults to "^.")`

	args, _ := docopt.ParseDoc(usage)

	site.SourcePath, _ = args.String("<source-path>")
	site.TargetPath, _ = args.String("<target-path>")
	site.includeMarkdown, _ = args.Bool("--include-markdown")
	site.includeTemplate, _ = args.Bool("--include-template")

	argIgnore, _ := args.String("--ignore")
	if argIgnore != "" {
		var err error
		site.copyOptions = new(fileutil.CopyOptions)
		site.copyOptions.IgnoreRegex, err = regexp.Compile(argIgnore)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := site.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := site.Build(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
