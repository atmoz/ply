package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atmoz/ply/fileutil"
)

const defaultTarget string = "ply.build"
const defaultIgnore string = `/\.`
const defaultFileMode os.FileMode = 0644
const defaultDirMode os.FileMode = 0755

type Site struct {
	Pages      []*Page
	SourcePath string
	TargetPath string
	Tags       map[string][]*Page

	plyPath         string
	includeMarkdown bool
	includeTemplate bool
	prettyUrls      bool
	keepLinks       bool
	copyOptions     *fileutil.CopyOptions

	templates map[string]*PlyTemplate
}

func (site *Site) Init() (err error) {
	if site.SourcePath == "" {
		if site.SourcePath, err = os.Getwd(); err != nil {
			fail(err)
		}
	}
	if site.SourcePath, err = filepath.Abs(site.SourcePath); err != nil {
		return err
	}

	if site.TargetPath == "" {
		site.TargetPath = filepath.Join(site.SourcePath, defaultTarget)
	}
	if site.TargetPath, err = filepath.Abs(site.TargetPath); err != nil {
		return err
	}

	if site.SourcePath == site.TargetPath {
		return errors.New("Target path can't be the same as source path")
	}

	if site.copyOptions == nil {
		site.copyOptions = new(fileutil.CopyOptions)
		site.copyOptions.IgnoreRegex = append(
			site.copyOptions.IgnoreRegex, regexp.MustCompile(defaultIgnore))
	}

	site.copyOptions.IgnoreRegex = append(
		site.copyOptions.IgnoreRegex, regexp.MustCompile("^"+site.TargetPath))

	if site.plyPath == "" {
		site.plyPath = filepath.Join(site.SourcePath, ".ply")
	}

	site.Tags = make(map[string][]*Page)
	site.templates = make(map[string]*PlyTemplate)
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
		if content, err := p.parse(); err == nil {
			if site.prettyUrls {
				if err = os.MkdirAll(p.AbsDir, defaultDirMode); err != nil {
					return err
				}
			}
			if err := ioutil.WriteFile(p.AbsPath, content, defaultFileMode); err != nil {
				return err
			}
			fmt.Println("Page:", p.AbsPath)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := filepath.Walk(site.TargetPath, site.cleanWalk); err != nil {
		return err
	}

	return nil
}

func (site *Site) buildWalk(path string, f os.FileInfo, err error) error {
	basename := filepath.Base(path)
	if strings.HasSuffix(basename, ".md") {
		if page, err := NewPage(site, path); err == nil {
			site.Pages = append(site.Pages, page)
		} else {
			return err
		}
	} else if basename == "ply.template" {
		if template, err := NewPlyTemplate(site, path); err != nil {
			return err
		} else {
			site.templates[filepath.Dir(path)] = template
		}
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
