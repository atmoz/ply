package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

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

	PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
