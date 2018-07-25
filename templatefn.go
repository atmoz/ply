package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"text/template"
)

func (site *Site) templateFnMap() template.FuncMap {
	return template.FuncMap{
		"pathDir":           filepath.Dir,
		"pathRel":           filepath.Rel,
		"pathMatch":         filepath.Match,
		"regexMatch":        RegexMatch,
		"regexReplaceAll":   RegexReplaceAll,
		"regexFind":         RegexFind,
		"regexFindSubmatch": RegexFindSubmatch,
		"include":           site.include,
	}
}

func (site *Site) include(path string) (string, error) {
	absPath := filepath.Join(site.TargetPath, path)
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func RegexMatch(pattern string, s string) (bool, error) {
	return regexp.MatchString(pattern, s)
}

func RegexReplaceAll(pattern string, src string, repl string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	} else {
		return re.ReplaceAllString(src, repl), nil
	}
}

func RegexFind(pattern string, s string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	} else {
		return re.FindString(s), nil
	}
}

func RegexFindSubmatch(pattern string, s string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	} else {
		return re.FindStringSubmatch(s), nil
	}
}
