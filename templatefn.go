package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
		"pathDir":           filepath.Dir,
		"pathRel":           filepath.Rel,
		"pathMatch":         filepath.Match,
		"regexMatch":        t.RegexMatch,
		"regexReplaceAll":   t.RegexReplaceAll,
		"regexFind":         t.RegexFind,
		"regexFindSubmatch": t.RegexFindSubmatch,
		"include":           t.Include,
		"templateImport":    t.TemplateImport,
		"templateWrite":     t.TemplateWrite,
		"yamlRead":          t.YamlRead,
		"yamlWrite":         t.YamlWrite,
		"stringsJoin":       t.Join,
		"array":             t.Array,
		"timeNow":           t.TimeNow,
		"timeFormat":        t.TimeFormat,
		"timeParse":         t.TimeParse,
	}
}

func (t *PlyTemplate) AbsRelToTemplate(path string) (string, error) {
	return fileutil.AbsRootLimit(t.site.TargetPath, filepath.Join(filepath.Dir(t.path), path))
}

func (t *PlyTemplate) Include(path string) (string, error) {
	absPath, err := t.AbsRelToTemplate(path)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (t *PlyTemplate) TemplateImport(path, name string) (string, error) {
	content, err := t.Include(path)
	if err != nil {
		return "", err
	}

	_, err = t.template.New(name).Parse(content)
	return "", err
}

func (t *PlyTemplate) TemplateWrite(name, path string, data interface{}) (string, error) {
	absPath, err := t.AbsRelToTemplate(path)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := t.template.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}

	fmt.Println("Write template", name, ":", absPath)
	return "", ioutil.WriteFile(absPath, buf.Bytes(), 0644)
}

func (t *PlyTemplate) YamlRead(path string) (data YamlData, err error) {
	absPath, err := t.AbsRelToTemplate(path)
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

func (t *PlyTemplate) YamlWrite(path string, data YamlData) (string, error) {
	absPath, err := t.AbsRelToTemplate(path)
	if err != nil {
		return "", err
	}

	content, err := yaml.Marshal(&data)
	if err != nil {
		return "", err
	}

	return "", ioutil.WriteFile(absPath, content, 0644)
}

func (t *PlyTemplate) RegexMatch(pattern string, s string) (bool, error) {
	return regexp.MatchString(pattern, s)
}

func (t *PlyTemplate) RegexReplaceAll(pattern string, src string, repl string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	} else {
		return re.ReplaceAllString(src, repl), nil
	}
}

func (t *PlyTemplate) RegexFind(pattern string, s string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	} else {
		return re.FindString(s), nil
	}
}

func (t *PlyTemplate) RegexFindSubmatch(pattern string, s string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	} else {
		return re.FindStringSubmatch(s), nil
	}
}

func (t *PlyTemplate) Join(a []interface{}, sep string) string {
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
