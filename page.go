package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

var reMarkdownHref *regexp.Regexp = regexp.MustCompile(`(<a[^>]*href=")([^"]+\.md)("[^>]*>)`)

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

func NewPage(site *Site, absPath string) (p *Page, err error) {
	if !filepath.IsAbs(absPath) {
		return nil, errors.New(absPath + " must be an absolute path!")
	}

	p = new(Page)
	p.Site = site
	p.AbsSrcPath = absPath
	p.AbsPath, p.Name = p.getTargetPathAndName(absPath)

	p.Path, err = filepath.Rel(site.TargetPath, p.AbsPath)
	if err != nil {
		return nil, err
	}

	p.AbsDir = filepath.Dir(p.AbsPath)
	p.Dir, err = filepath.Rel(site.TargetPath, p.AbsDir)
	if err != nil {
		return nil, err
	}

	p.DirParts = make(map[string]string)
	dirNames := strings.Split(p.Dir, string(filepath.Separator))
	for i, v := range dirNames {
		dirPath := filepath.Join(dirNames[0:i]...)
		p.DirParts[filepath.Join(dirPath, v)] = v
	}

	content, err := ioutil.ReadFile(p.AbsSrcPath)
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
		p.Title = p.Name
	}

	if err := p.registerTags(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Page) getTargetPathAndName(path string) (newPath string, name string) {
	nameFromBase := func(path string) string {
		return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	if filepath.Base(path) == "index.md" { // index.md -> index.html
		newPath = strings.Replace(path, ".md", ".html", 1)
		name = nameFromBase(newPath)
	} else if strings.HasSuffix(path, ".html.md") { // path.html.md -> path.html
		newPath = strings.Replace(path, ".md", "", 1)
		name = nameFromBase(newPath)
	} else if p.Site.prettyUrls { // path.md -> path/index.html
		newPath = filepath.Join(strings.Replace(path, ".md", "", 1), "index.html")
		name = filepath.Dir(newPath)
	} else { // path.md -> path.html
		newPath = strings.Replace(path, ".md", ".html", 1)
		name = nameFromBase(newPath)
	}

	return
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

func (p *Page) Rel(path string) (string, error) {
	return filepath.Rel(p.AbsDir, filepath.Join(p.Site.TargetPath, path))
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

	// Replace internal .md links
	if !p.Site.keepLinks {
		for _, match := range reMarkdownHref.FindAllSubmatch(content, -1) {
			href := string(match[2])

			hrefUrl, err := url.Parse(href)
			if err != nil {
				continue // Ignore, this may not be a normal URL
			}

			if hrefUrl.IsAbs() {
				continue // Ignore, do not touch absolute URL
			}

			before := string(match[1])
			href = strings.TrimSuffix(href, ".md")
			if !p.Site.prettyUrls {
				href += ".html"
			}
			after := string(match[3])
			content = bytes.Replace(content, match[0], []byte(before+href+after), -1)
		}
	}

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
