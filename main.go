package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docopt/docopt-go"
	"gopkg.in/russross/blackfriday.v2"
)

type Site []Page

type Page struct {
	Sitemap    *Site
	Title      string `xml:"h1"`
	Content    string
	Path       string
	SourcePath string
}

var templates map[string]*template.Template
var sitemap Site

func main() {
	usage := `ply - recursive markdown to HTML converter

Usage: ply [clean] [<path>] [-h|--help]`

	args, _ := docopt.ParseDoc(usage)
	argClean, _ := args.Bool("clean")
	argPath, _ := args.String("<path>")

	templates = make(map[string]*template.Template)

	if argPath == "" {
		argPath = "."
	} else {
		argPath = filepath.Clean(argPath)
	}

	var err error
	if argClean {
		err = filepath.Walk(argPath, clean)
	} else {
		err = filepath.Walk(argPath, build)
	}

	if err != nil {
		panic(err)
	}

	for _, p := range sitemap {
		// Apply templates recursively
		dirname := filepath.Dir(p.Path)
		for {
			if templates[dirname] != nil {
				var content bytes.Buffer
				templates[dirname].Execute(&content, p)
				p.Content = content.String()
			}

			if dirname == "." {
				break
			}

			dirname = filepath.Dir(dirname) // Remove last element
		}

		ioutil.WriteFile(p.Path, []byte(p.Content), 0644)
		fmt.Println("Created:", p.Path)
	}
}

func build(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".md") {
		htmlPath := strings.Replace(path, ".md", ".html", 1)

		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		htmlContent := blackfriday.Run(content)

		var p Page
		err = xml.Unmarshal([]byte("<r>"+string(htmlContent)+"</r>"), &p)
		if err != nil || p.Title == "" {
			p.Title = htmlPath
		}

		p.Sitemap = &sitemap
		p.Content = string(htmlContent)
		p.Path = htmlPath
		p.SourcePath = path

		sitemap = append(sitemap, p)
	} else if filepath.Base(path) == "ply.template" {
		dirname := filepath.Dir(path)
		templates[dirname] = template.Must(template.ParseFiles(path))
	}
	return nil
}

func clean(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".md") {
		htmlPath := strings.Replace(path, ".md", ".html", 1)
		if err := os.Remove(htmlPath); err != nil {
			fmt.Println("Failed to remove:", htmlPath, err)
		} else {
			fmt.Println("Removed:", htmlPath)
		}
	}
	return nil
}
