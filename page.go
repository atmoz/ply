package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	site     *Site
	Title    string
	Content  string
	Dir      string
	AbsDir   string
	DirParts map[string]string
	Path     string
	AbsPath  string
}

func NewPage(site *Site, path string) *Page {
	p := new(Page)
	p.site = site
	p.AbsDir = filepath.Dir(path)
	p.Dir, _ = filepath.Rel(site.TargetPath, p.AbsDir)
	p.AbsPath = strings.Replace(path, ".md", ".html", 1)
	p.Path, _ = filepath.Rel(site.TargetPath, p.AbsPath)

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

	reTitle := regexp.MustCompile(`<h[1-6]>(.*)</h[1-6]>`)
	match := reTitle.FindStringSubmatch(p.Content)
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

func (p *Page) SitemapReversed() []*Page {
	reversed := make([]*Page, len(p.site.Pages))
	for i := range p.site.Pages {
		reversed[i] = p.site.Pages[len(p.site.Pages)-1-i]
	}
	return reversed
}

func (p *Page) SiteRoot() string {
	path, _ := filepath.Rel(p.AbsDir, p.site.TargetPath)
	return filepath.Clean(path)
}
