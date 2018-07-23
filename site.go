package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Site struct {
	Pages []*Page
	Path  string

	templates     map[string]*template.Template
	templateFnMap template.FuncMap
}

func (site *Site) Init() {
	if absPath, err := filepath.Abs(site.Path); err != nil {
		panic(err)
	} else {
		site.Path = absPath
	}

	site.templates = make(map[string]*template.Template)
	site.templateFnMap = TemplateFnMap()
}

func (site *Site) Build() error {
	if err := filepath.Walk(site.Path, site.buildWalk); err != nil {
		return err
	}

	for _, p := range site.Pages {
		// Apply templates recursively
		dirname := filepath.Dir(p.AbsPath)
		for {
			if site.templates[dirname] != nil {
				var content bytes.Buffer
				if err := site.templates[dirname].Execute(&content, p); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				p.Content = content.String()
			}

			if rel, err := filepath.Rel(site.Path, dirname); err != nil {
				panic(err)
			} else if rel == site.Path || rel == "." || rel == ".." || rel == "../.." {
				break
			}

			dirname = filepath.Dir(dirname) // Remove last element
		}

		ioutil.WriteFile(p.AbsPath, []byte(p.Content), 0644)
		fmt.Println("Created:", p.AbsPath)
	}
	return nil
}

func (site *Site) buildWalk(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".md") {
		site.Pages = append(site.Pages, NewPage(site, path))
	} else if filepath.Base(path) == "ply.template" {
		dirname := filepath.Dir(path)
		tp := template.New(path).Funcs(site.templateFnMap)
		templateContent, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		tp, err = tp.Parse(string(templateContent))
		if err != nil {
			return err
		}
		site.templates[dirname] = tp
	}
	return nil
}

func (site *Site) Clean() error {
	return filepath.Walk(site.Path, site.cleanWalk)
}

func (site *Site) cleanWalk(path string, f os.FileInfo, err error) error {
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
