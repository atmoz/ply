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
	Name       string
	Dir        string
	AbsDir     string
	DirParts   map[string]string
	Path       string
	AbsPath    string
	AbsSrcPath string
	Meta       map[string]interface{}
	tags       []string

	content []byte
}

func NewPage(site *Site, path string) (*Page, error) {
	p := new(Page)
	p.Site = site
	p.AbsDir = filepath.Dir(path)
	p.Dir, _ = filepath.Rel(site.TargetPath, p.AbsDir)
	p.AbsPath = strings.Replace(path, ".md", ".html", 1)
	p.Path, _ = filepath.Rel(site.TargetPath, p.AbsPath)
	p.AbsSrcPath = path
	p.Name = filepath.Base(p.Path)

	p.DirParts = make(map[string]string)
	dirNames := strings.Split(p.Dir, string(filepath.Separator))
	for i, v := range dirNames {
		dirPath := filepath.Join(dirNames[0:i]...)
		p.DirParts[filepath.Join(dirPath, v)] = v
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
	rel, _ := filepath.Rel(p.AbsDir, p.Site.TargetPath)
	return rel
}

func (p *Page) Rel(path string) (rel string, err error) {
	return filepath.Rel(p.AbsDir, filepath.Clean(filepath.Join(p.Site.TargetPath, path)))
}

func (p *Page) Content() (content string, err error) {
	bytes, err := p.ContentBytes()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (p *Page) ContentBytes() (content []byte, err error) {
	if p.content != nil {
		return p.content, nil
	}

	content, err = ioutil.ReadFile(p.AbsSrcPath)
	if err != nil {
		return nil, err
	}

	_, content, err = splitMetaAndContent(content)
	if err != nil {
		return nil, err
	}

	content = blackfriday.Run(content)
	return content, nil
}

func (p *Page) parse() (result []byte, err error) {
	p.content, err = p.ContentBytes()
	if err != nil {
		return nil, err
	}

	// Apply templates recursively
	dirname := filepath.Dir(p.AbsPath)
	for {
		if p.Site.templates[dirname] != nil {
			var templateBuffer bytes.Buffer
			if err := p.Site.templates[dirname].template.Execute(&templateBuffer, p); err != nil {
				return nil, err
			}
			p.content = templateBuffer.Bytes() // Update content from template
		}

		// Break loop when we are on root (last) level
		if rel, err := filepath.Rel(p.Site.TargetPath, dirname); err != nil {
			panic(err)
		} else if rel == "." || rel == "" {
			break
		}

		dirname = filepath.Dir(dirname) // Remove last dir
	}

	result = p.content
	p.content = nil // This was only needed for templating
	return result, nil
}

func (p *Page) registerTags() error {
	if p.Meta["tags"] == nil {
		return nil
	}

	if t := reflect.TypeOf(p.Meta["tags"]).String(); t != "[]interface {}" {
		return errors.New(p.Path + ": metadata \"tags\" must be of type []string, but was " + t)
	}

	tags := p.Meta["tags"].([]interface{})
	for _, tag := range tags {
		if t := reflect.TypeOf(tag).String(); t != "string" {
			return errors.New(p.Path + ": tag must be of type string, but was " + t)
		}
		name := tag.(string)
		p.tags = append(p.tags, name)
		site.Tags[name] = append(site.Tags[name], p)
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
