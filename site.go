package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/atmoz/ply/fileutil"
)

type Site struct {
	Pages      []*Page
	SourcePath string
	TargetPath string
	Tags       map[string][]*Page

	plyPath         string
	includeMarkdown bool
	includeTemplate bool
	copyOptions     *fileutil.CopyOptions

	templates map[string]*template.Template
}

func (site *Site) Init() (err error) {
	if site.SourcePath, err = filepath.Abs(site.SourcePath); err != nil {
		return err
	}
	if site.TargetPath, err = filepath.Abs(site.TargetPath); err != nil {
		return err
	}

	if site.SourcePath != site.TargetPath && strings.HasPrefix(site.TargetPath, site.SourcePath) {
		return errors.New("Target path can't be a under source path")
	}
	if site.copyOptions == nil {
		site.copyOptions = new(fileutil.CopyOptions)
		site.copyOptions.IgnoreRegex = regexp.MustCompile("^\\.") // Hidden files
	}

	if site.plyPath == "" {
		site.plyPath = filepath.Join(site.SourcePath, ".ply")
	}

	site.Tags = make(map[string][]*Page)
	site.templates = make(map[string]*template.Template)
	return nil
}

func (site *Site) Build() error {
	if site.SourcePath != site.TargetPath {
		if err := fileutil.CopyDirectory(site.SourcePath, site.TargetPath, site.copyOptions); err != nil {
			return err
		}

		if info, err := os.Stat(site.plyPath); !os.IsNotExist(err) && info.IsDir() {
			if err := fileutil.CopyDirectory(site.plyPath, site.TargetPath, nil); err != nil {
				return err
			}
		}
	}

	if err := filepath.Walk(site.TargetPath, site.buildWalk); err != nil {
		return err
	}

	for _, p := range site.Pages {
		if content, err := p.parse(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			ioutil.WriteFile(p.AbsPath, content, 0644)
			fmt.Println("Created:", p.AbsPath)
		}
	}

	if err := filepath.Walk(site.TargetPath, site.cleanWalk); err != nil {
		return err
	}

	return nil
}

func (site *Site) buildWalk(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".md") {
		if page, err := NewPage(site, path); err == nil {
			site.Pages = append(site.Pages, page)
		} else {
			return err
		}
	} else if filepath.Base(path) == "ply.template" {
		dirname := filepath.Dir(path)
		tp := template.New(path).Funcs(site.templateFnMap())
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

func (site *Site) cleanWalk(path string, f os.FileInfo, err error) error {
	cleanMarkdown := !site.includeMarkdown && strings.HasSuffix(path, ".md")
	cleanTemplate := !site.includeTemplate && filepath.Base(path) == "ply.template"
	if cleanMarkdown || cleanTemplate {
		os.Remove(path)
	}
	return nil
}
