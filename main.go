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
  ply [options] [<source-path>] [<target-path>]
  ply -h|--help

Options:
  --include-markdown    Include markdown files in target
  --include-template    Include template files in target
  --pretty-urls         Use <path>/index.html trick for pretty urls
  --keep-links          Do NOT replace internal *.md links with *.html
  --ignore=<regex>      File names to ignore (defaults to "/\.")
  --allow-exec          Allow templates to execute commands (BE CAREFUL!)
  `

	args, _ := docopt.ParseDoc(usage)

	site.SourcePath, _ = args.String("<source-path>")
	site.TargetPath, _ = args.String("<target-path>")
	site.includeMarkdown, _ = args.Bool("--include-markdown")
	site.includeTemplate, _ = args.Bool("--include-template")
	site.prettyUrls, _ = args.Bool("--pretty-urls")
	site.keepLinks, _ = args.Bool("--keep-links")
	site.allowExec, _ = args.Bool("--allow-exec")

	argIgnore, _ := args.String("--ignore")
	if argIgnore != "" {
		var err error
		site.copyOptions = new(fileutil.CopyOptions)
		re, err := regexp.Compile(argIgnore)
		if err != nil {
			fail(err)
		}
		site.copyOptions.IgnoreRegex = append(site.copyOptions.IgnoreRegex, re)
	}

	if err := site.Init(); err != nil {
		fail(err)
	}

	if err := site.Build(); err != nil {
		fail(err)
	}

	fmt.Println(len(site.Pages), "pages")
	PrintMemUsage()
}

func fail(err error) {
	fmt.Println("FATAL:", err)
	os.Exit(1)
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
