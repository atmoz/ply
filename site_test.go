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

func compareFiles(a, b string) bool {
	contentA, _ := ioutil.ReadFile(a)
	contentB, _ := ioutil.ReadFile(b)

	if string(contentA) == "" || string(contentA) != string(contentB) {
		fmt.Println("Files are not equal:", a, "and", b)
		fmt.Println(a, string(contentA))
		fmt.Println(b, string(contentB))
		return false
	}

	return true
}

func compareAllExpectedFiles(site *Site) bool {
	numFiles := 0
	walkFn := func(path string, info os.FileInfo, err error) error {
		testPath := filepath.Clean(strings.Replace(path, "ply.expected", "", 1))
		if info != nil && !info.IsDir() {
			numFiles++
			if !compareFiles(path, testPath) {
				return errors.New(testPath + " was not as expected")
			}
		}
		return nil
	}

	if err := filepath.Walk(filepath.Join(site.TargetPath, "ply.expected"), walkFn); err != nil {
		fmt.Println(err)
		return false
	}

	if numFiles == 0 {
		fmt.Println("Found no files to compare")
		return false
	}

	return true
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

func TestExampleBlog(t *testing.T) {
	var site Site
	if !buildAndCompare(&site, "example_blog") {
		t.Fail()
	}
}
