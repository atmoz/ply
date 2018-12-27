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

type PageMeta map[string]interface{}

type Page struct {
	Site  *Site
	Title string
	Name  string
	Path  *Path
	Meta  PageMeta
	Data  interface{}
	tags  []string

	content []byte
}

func NewPage(site *Site, absSrcPath string) (p *Page, err error) {
	p = new(Page)
	err = p.init(site, absSrcPath)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(p.Path.AbsSrc)
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

func (p *Page) init(site *Site, absSrcPath string) (err error) {
	if !filepath.IsAbs(absSrcPath) {
		return errors.New(absSrcPath + " must be an absolute path!")
	}

	p.Site = site
	p.Path, err = NewPath(site, absSrcPath)
	if err != nil {
		return err
	}

	if filepath.Base(p.Path.Rel) == "index.html" {
		p.Name = filepath.Base(p.Path.Rel)
	} else {
		p.Name = strings.TrimSuffix(filepath.Base(p.Path.Rel), filepath.Ext(p.Path.Rel))
	}

	return nil
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

	content, err = ioutil.ReadFile(p.Path.AbsSrc)
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
	dirname := p.Path.AbsDir
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
		return errors.New(p.Path.Rel + ": metadata \"tags\" must be of type []string, but was " + t)
	}

	tags := p.Meta["tags"].([]interface{})
	for _, tag := range tags {
		if t := reflect.TypeOf(tag).String(); t != "string" {
			return errors.New(p.Path.Rel + ": tag must be of type string, but was " + t)
		}
		name := tag.(string)
		p.tags = append(p.tags, name)
		site.Tags[name] = append(site.Tags[name], p)
	}

	return nil
}

func (p *Page) Url() string {
	return p.Path.Url()
}

func (p *Page) UrlDirParts() map[string]string {
	return p.Path.UrlDirParts()
}

func (p *Page) UrlRelTo(relPath string) (string, error) {
	return p.Path.UrlRelTo(relPath)
}

func (p *Page) UrlToRoot() string {
	return p.Path.UrlToRoot()
}

func splitMetaAndContent(content []byte) (PageMeta, []byte, error) {
	re := regexp.MustCompile(`(?ms:\A-{3,}\s*(.+?)-{3,}\s*(.*)\z)`)
	match := re.FindSubmatch(content)

	if len(match) != 3 {
		return nil, content, nil
	}

	var meta PageMeta
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

type EmptyPage struct {
	Page
}

func NewEmptyPage(site *Site, absPath string) (p *EmptyPage, err error) {
	p = new(EmptyPage)
	err = p.init(site, absPath)
	if err != nil {
		return nil, err
	}

	p.Title = p.Name
	return p, nil
}

func (p *EmptyPage) Content() (content string, err error) {
	return "", nil
}

func (p *EmptyPage) ContentBytes() (content []byte, err error) {
	return []byte(""), nil
}
