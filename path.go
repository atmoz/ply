package main

import (
	"errors"
	"path/filepath"
	"strings"
)

type Path struct {
	site      *Site
	Rel       string
	Abs       string
	AbsSrc    string
	RelDir    string
	AbsDir    string
	DirParts  map[string]string
	RelToRoot string
}

func NewPath(site *Site, absSrcPath string) (p *Path, err error) {
	p = new(Path)

	if !filepath.IsAbs(absSrcPath) {
		return nil, errors.New(absSrcPath + " must be an absolute path!")
	}

	p.site = site

	p.AbsSrc = absSrcPath
	p.Abs = p.getTargetPath(p.AbsSrc)

	p.Rel, err = filepath.Rel(site.TargetPath, p.Abs)
	if err != nil {
		return nil, err
	}

	p.AbsDir = filepath.Dir(p.Abs)
	p.RelDir, err = filepath.Rel(site.TargetPath, p.AbsDir)
	if err != nil {
		return nil, err
	}

	p.DirParts = make(map[string]string)
	dirNames := strings.Split(p.RelDir, string(filepath.Separator))
	for i, v := range dirNames {
		dirPath := filepath.Join(dirNames[0:i]...)
		p.DirParts[filepath.Join(dirPath, v)] = v
	}

	p.RelToRoot, _ = filepath.Rel(p.AbsDir, p.site.TargetPath)

	return p, nil
}

func (p *Path) getTargetPath(path string) string {
	if filepath.Base(path) == "index.md" { // index.md -> index.html
		return strings.Replace(path, ".md", ".html", 1)
	} else if strings.HasSuffix(path, ".html.md") { // path.html.md -> path.html
		return strings.Replace(path, ".md", "", 1)
	} else if p.site.prettyUrls { // path.md -> path/index.html
		return filepath.Join(strings.Replace(path, ".md", "", 1), "index.html")
	} else { // path.md -> path.html
		return strings.Replace(path, ".md", ".html", 1)
	}
}

func (p *Path) RelTo(path string) (string, error) {
	return filepath.Rel(p.AbsDir, filepath.Join(p.site.TargetPath, path))
}

func (p *Path) String() string {
	// Normalize for use in templates
	return strings.Replace(p.Rel, string(filepath.Separator), "/", -1)
}

func (p *Path) Url() string {
	return normalizePathToUrl(p.Rel)
}

func (p *Path) UrlDirParts() map[string]string {
	urlParts := make(map[string]string)
	pathParts := p.DirParts
	for path, name := range pathParts {
		urlParts[normalizePathToUrl(path)] = name
	}

	return urlParts
}

func (p *Path) UrlRelTo(to string) (string, error) {
	path, err := p.RelTo(to)
	if err != nil {
		return "", err
	}

	return normalizePathToUrl(path), nil
}

func (p *Path) UrlToRoot() string {
	return normalizePathToUrl(p.RelToRoot)
}

func normalizePathToUrl(path string) string {
	if filepath.Separator != '/' {
		return strings.Replace(path, string(filepath.Separator), "/", -1)
	} else {
		return path
	}
}
