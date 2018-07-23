package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	site       *Site
	Title      string `xml:"h1"`
	Content    string
	Dir        string
	AbsDir     string
	DirParts   map[string]string
	Path       string
	AbsPath    string
	SrcPath    string
	AbsSrcPath string
}

func NewPage(site *Site, path string) *Page {
	p := new(Page)
	p.site = site
	p.AbsDir = filepath.Dir(path)
	p.Dir, _ = filepath.Rel(site.Path, p.AbsDir)
	p.AbsPath = strings.Replace(path, ".md", ".html", 1)
	p.Path, _ = filepath.Rel(site.Path, p.AbsPath)
	p.AbsSrcPath = path
	p.SrcPath, _ = filepath.Rel(site.Path, p.AbsSrcPath)

	p.DirParts = make(map[string]string)
	dirNames := strings.Split(p.Dir, string(filepath.Separator))
	for i, v := range dirNames {
		dirPath := strings.Join(dirNames[0:i], string(filepath.Separator))
		p.DirParts[filepath.Join(p.SiteRoot(), dirPath, v)] = v
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	htmlContent := blackfriday.Run(content)
	p.Content = string(htmlContent)

	re := regexp.MustCompile(`<h[1-6]>(.*)</h[1-6]>`)
	match := re.FindStringSubmatch(p.Content)
	if match != nil {
		p.Title = match[1]
	} else {
		p.Title = p.Path
	}

	return p
}

func (p *Page) Sitemap() []*Page {
	return p.site.Pages
}

func (p *Page) SiteRoot() string {
	path, _ := filepath.Rel(p.AbsDir, p.site.Path)
	return filepath.Clean(path)
}
