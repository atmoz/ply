package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	urlpath "path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/atmoz/ply/fileutil"
	yaml "gopkg.in/yaml.v2"
)

type YamlData map[string]interface{}

type PlyTemplate struct {
	path     string
	site     *Site
	template *template.Template
}

func NewPlyTemplate(site *Site, path string) (t *PlyTemplate, err error) {
	t = &PlyTemplate{}
	t.path = path
	t.site = site
	t.template = template.New(path).Funcs(t.templateFnMap())
	templateContent, err := ioutil.ReadFile(path)
	if err != nil {
		return t, err
	}

	_, err = t.template.Parse(string(templateContent))
	return t, err
}

func (t *PlyTemplate) templateFnMap() template.FuncMap {
	return template.FuncMap{
		"urlBase":           urlpath.Base,
		"urlDir":            urlpath.Dir,
		"urlExt":            urlpath.Ext,
		"urlMatch":          urlpath.Match,
		"urlRel":            urlRel,
		"listDirs":          t.ListDirs,
		"listFiles":         t.ListFiles,
		"hasPage":           t.HasPage,
		"hasFile":           t.HasFile,
		"hasFileOrPage":     t.HasFileOrPage,
		"regexMatch":        t.RegexMatch,
		"regexReplaceAll":   t.RegexReplaceAll,
		"regexFind":         t.RegexFind,
		"regexFindSubmatch": t.RegexFindSubmatch,
		"include":           t.Include,
		"templateImport":    t.TemplateImport,
		"templateWrite":     t.TemplateWrite,
		"yamlRead":          t.YamlRead,
		"yamlWrite":         t.YamlWrite,
		"stringsJoin":       t.StringsJoin,
		"stringsSplit":      strings.Split,
		"array":             t.Array,
		"timeNow":           t.TimeNow,
		"timeFormat":        t.TimeFormat,
		"timeParse":         t.TimeParse,
		"mathInc":           t.MathInc,
		"mathDec":           t.MathDec,
	}
}

func (t *PlyTemplate) AbsRelToTemplate(url string) (string, error) {
	return fileutil.AbsRootLimit(t.site.TargetPath, filepath.Join(filepath.Dir(t.path), url))
}

func (t *PlyTemplate) ListDirs(url string, recursive bool) (map[string]string, error) {
	return t.listFiles(url, true, recursive)
}

func (t *PlyTemplate) ListFiles(url string, recursive bool) (map[string]string, error) {
	return t.listFiles(url, false, recursive)
}

func (t *PlyTemplate) listFiles(url string, dirNotFile bool, recursive bool) (map[string]string, error) {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return nil, err
	}

	list := make(map[string]string)
	walkFn := func(subpath string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(absPath, subpath)
		if err != nil {
			return err
		}

		// Ignore sub directories
		if !recursive && len(strings.Split(relPath, "/")) > 1 {
			return nil
		}

		// Dir or file
		if (dirNotFile && !info.IsDir()) || (!dirNotFile && info.IsDir()) {
			return nil
		}

		// Filter out pages and templates
		ext := filepath.Ext(relPath)
		if relPath == "." || ext == ".md" || ext == ".html" || ext == ".template" {
			return nil
		}

		list[normalizePathToUrl(relPath)] = filepath.Base(relPath)
		return nil
	}

	if err := filepath.Walk(absPath, walkFn); err != nil {
		return nil, err
	}

	return list, nil
}

func (t *PlyTemplate) HasPage(url string) bool {
	for _, p := range t.site.Pages {
		if p.Path.Url() == url {
			return true
		}
	}

	return false
}

func (t *PlyTemplate) HasFile(url string) bool {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	return os.IsExist(err)
}

func (t *PlyTemplate) HasFileOrPage(url string) bool {
	return t.HasFile(url) || t.HasPage(url)
}

func (t *PlyTemplate) Include(url string) (string, error) {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (t *PlyTemplate) TemplateImport(url, name string) (string, error) {
	content, err := t.Include(url)
	if err != nil {
		return "", err
	}

	_, err = t.template.New(name).Parse(content)
	return "", err
}

func (t *PlyTemplate) TemplateWrite(name, url string, data interface{}) (string, error) {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return "", err
	}

	page, err := NewEmptyPage(t.site, absPath)
	if err != nil {
		return "", err
	}
	page.Data = data

	buf := &bytes.Buffer{}
	if err := t.template.ExecuteTemplate(buf, name, page); err != nil {
		return "", err
	}

	fmt.Println("File from template", name+":", absPath)
	return "", ioutil.WriteFile(absPath, buf.Bytes(), 0644)
}

func (t *PlyTemplate) YamlRead(url string) (data YamlData, err error) {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return data, err
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return data, err
	}

	err = yaml.Unmarshal(content, &data)
	return data, err
}

func (t *PlyTemplate) YamlWrite(url string, data YamlData) (string, error) {
	absPath, err := t.AbsRelToTemplate(url)
	if err != nil {
		return "", err
	}

	content, err := yaml.Marshal(&data)
	if err != nil {
		return "", err
	}

	return "", ioutil.WriteFile(absPath, content, 0644)
}

var regexCache map[string]*regexp.Regexp

func (t *PlyTemplate) RegexCompileCache(pattern string) (*regexp.Regexp, error) {
	if regexCache == nil {
		regexCache = make(map[string]*regexp.Regexp)
	}

	var err error
	if regexCache[pattern] == nil {
		regexCache[pattern], err = regexp.Compile(pattern)
	}

	return regexCache[pattern], err
}

func (t *PlyTemplate) RegexMatch(pattern string, s string) (bool, error) {
	re, err := t.RegexCompileCache(pattern)
	if err != nil {
		return false, err
	} else {
		return re.MatchString(s), nil
	}
}

func (t *PlyTemplate) RegexReplaceAll(pattern string, src string, repl string) (string, error) {
	re, err := t.RegexCompileCache(pattern)
	if err != nil {
		return "", err
	} else {
		return re.ReplaceAllString(src, repl), nil
	}
}

func (t *PlyTemplate) RegexFind(pattern string, s string) (string, error) {
	re, err := t.RegexCompileCache(pattern)
	if err != nil {
		return "", err
	} else {
		return re.FindString(s), nil
	}
}

func (t *PlyTemplate) RegexFindSubmatch(pattern string, s string) ([]string, error) {
	re, err := t.RegexCompileCache(pattern)
	if err != nil {
		return nil, err
	} else {
		return re.FindStringSubmatch(s), nil
	}
}

func (t *PlyTemplate) StringsJoin(a []interface{}, sep string) string {
	ss := make([]string, len(a))
	for i, v := range a {
		ss[i] = v.(string)
	}
	return strings.Join(ss, sep)
}

func (t *PlyTemplate) Array(ss ...interface{}) []interface{} {
	return ss
}

func (t *PlyTemplate) TimeNow() time.Time {
	return time.Now()
}

func (t *PlyTemplate) TimeFormat(u time.Time, layout string) string {
	return u.Format(layout)
}

func (t *PlyTemplate) TimeParse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

func (t *PlyTemplate) MathInc(i int) int {
	return i + 1
}

func (t *PlyTemplate) MathDec(i int) int {
	return i - 1
}

func urlRel(baseUrl, targetUrl string) (string, error) {
	rel, err := filepath.Rel(baseUrl, targetUrl)
	if err != nil {
		return "", err
	}

	return normalizePathToUrl(rel), nil
}
