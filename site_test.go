package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atmoz/ply/fileutil"
)

func copyTestDir(subdir string) string {
	dir, err := ioutil.TempDir("", "ply")
	if err != nil {
		panic(err)
	}

	fileutil.CopyDirectory(filepath.Join("test_files", subdir), dir, nil)
	return dir
}

func compareWithExpected(file string) bool {
	content, _ := ioutil.ReadFile(file)

	expectedFile := filepath.Join(file + ".expected")
	expectedContent, _ := ioutil.ReadFile(expectedFile)

	if string(content) == "" || string(content) != string(expectedContent) {
		fmt.Println("test", file, string(content))
		fmt.Println("expe", expectedFile, string(expectedContent))
		fmt.Println("File", file, "was not as expected")
		return false
	}

	return true
}

func compareAllExpectedFiles(site *Site) bool {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".expected" {
			testPath := strings.TrimSuffix(path, filepath.Ext(path))
			if !compareWithExpected(testPath) {
				return errors.New(testPath + " was not as expected")
			}
		}
		return nil
	}

	return filepath.Walk(site.TargetPath, walkFn) == nil
}

func buildAndCompare(site *Site, path string) bool {
	site.SourcePath = copyTestDir(path)
	defer os.RemoveAll(site.TargetPath)

	if err := site.Init(); err != nil {
		fmt.Println(err)
		return false
	}

	site.Build()
	return compareAllExpectedFiles(site)
}

func TestOnePage(t *testing.T) {
	var site Site
	if !buildAndCompare(&site, "one_page") {
		t.Fail()
	}
}

func TestOneTemplate(t *testing.T) {
	var site Site
	if !buildAndCompare(&site, "one_template") {
		t.Fail()
	}
}

func TestLinks(t *testing.T) {
	var site Site
	if !buildAndCompare(&site, "links") {
		t.Fail()
	}
}

func TestLinksPretty(t *testing.T) {
	var site Site
	site.prettyUrls = true
	if !buildAndCompare(&site, "links_pretty") {
		t.Fail()
	}
}
