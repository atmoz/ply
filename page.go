package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

type Page struct {
	Site       *Site
	Title      string
	Content    string
	Dir        string
	AbsDir     string
	DirParts   map[string]string
	Path       string
	AbsPath    string
	AbsSrcPath string
	Meta       map[string]interface{}
}

func NewPage(site *Site, path string) (*Page, error) {
	p := new(Page)
	p.Site = site
	p.AbsDir = filepath.Dir(path)
	p.Dir, _ = filepath.Rel(site.TargetPath, p.AbsDir)
	p.AbsPath = strings.Replace(path, ".md", ".html", 1)
	p.Path, _ = filepath.Rel(site.TargetPath, p.AbsPath)
	p.AbsSrcPath = path

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

	p.Meta, content, err = splitMetaAndContent(content)
	if err != nil {
		panic(err)
	}

	if p.Meta["title"] != nil {
		p.Title = p.Meta["title"].(string)
	} else if h := findFirstHeading(content); h != "" {
		p.Title = h
	} else {
		p.Title = p.Path
	}

	if err := p.registerTags(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Page) Sitemap() []*Page {
	return p.Site.Pages
}

func (p *Page) SitemapReversed() []*Page {
	reversed := make([]*Page, len(p.Site.Pages))
	for i := range p.Site.Pages {
		reversed[i] = p.Site.Pages[len(p.Site.Pages)-1-i]
	}
	return reversed
}

func (p *Page) SiteRoot() string {
	path, _ := filepath.Rel(p.AbsDir, p.Site.TargetPath)
	return filepath.Clean(path)
}

func (p *Page) parse() (result []byte, err error) {
	// Markdown to HTML

	result, err = ioutil.ReadFile(p.AbsSrcPath)
	if err != nil {
		return nil, err
	}

	_, result, err = splitMetaAndContent(result)
	if err != nil {
		return nil, err
	}

	p.Content = string(blackfriday.Run(result))

	// Apply templates recursively
	dirname := filepath.Dir(p.AbsPath)
	for {
		if p.Site.templates[dirname] != nil {
			var content bytes.Buffer
			if err := p.Site.templates[dirname].Execute(&content, p); err != nil {
				return nil, err
			}
			p.Content = string(content.Bytes())
		}

		// Break loop when we are on root (last) level
		if rel, err := filepath.Rel(p.Site.TargetPath, dirname); err != nil {
			panic(err)
		} else if rel == "." || rel == "" {
			break
		}

		dirname = filepath.Dir(dirname) // Remove last dir
	}

	result = []byte(p.Content)
	p.Content = "" // No longer needed
	return result, nil
}

func (p *Page) registerTags() error {
	if p.Meta["tags"] != nil {
		if t := reflect.TypeOf(p.Meta["tags"]).String(); t != "[]interface {}" {
			return errors.New(p.Path + ": metadata \"tags\" must be of type []string, but was " + t)
		}
		tags := p.Meta["tags"].([]interface{})
		for _, tag := range tags {
			if t := reflect.TypeOf(tag).String(); t != "string" {
				return errors.New(p.Path + ": tag must be of type string, but was " + t)
			}
			name := tag.(string)
			site.Tags[name] = append(site.Tags[name], p)
		}
	}
	return nil
}

func splitMetaAndContent(content []byte) (map[string]interface{}, []byte, error) {
	re := regexp.MustCompile(`(?ms:\A-{3,}\s*(.+?)-{3,}\s*(.*)\z)`)
	match := re.FindSubmatch(content)

	if len(match) != 3 {
		return nil, content, nil
	}

	var meta map[string]interface{}
	metaRaw := match[1]
	content = match[2]

	if err := yaml.Unmarshal(metaRaw, &meta); err != nil {
		return nil, content, err
	}

	return meta, content, nil
}

func findFirstHeading(content []byte) string {
	re := regexp.MustCompile(`\s*#{1,6}\s*(.+)|\s*(.+)\s*\n\s*[-=]{1,}\s*`)
	match := re.FindSubmatch(content)

	if len(match) == 3 {
		return string(append(match[1], match[2]...)) // Only one group is captured
	} else {
		return ""
	}
}