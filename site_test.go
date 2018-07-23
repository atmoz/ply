package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/atmoz/ply/fileutil"
)

func copyTestDir(subdir string) string {
	dir, err := ioutil.TempDir("", "ply")
	if err != nil {
		panic(err)
	}

	fileutil.CopyDirectory(filepath.Join("test_files", subdir), dir)
	return dir
}

func compareWithExpected(file string) bool {
	content, _ := ioutil.ReadFile(file)

	expectedFile := filepath.Join(file + ".expected")
	expectedContent, _ := ioutil.ReadFile(expectedFile)

	if string(content) == "" || string(content) != string(expectedContent) {
		//fmt.Println("test", file, string(content))
		//fmt.Println("expe", expectedFile, string(expectedContent))
		fmt.Println("File", file, "was not as expected")
		return false
	}

	return true
}

func build(dir string) (site Site) {
	site.Path = dir
	site.Init()
	site.Build()
	return
}

func TestOnePage(t *testing.T) {
	dir := copyTestDir("one_page")
	defer os.RemoveAll(dir)
	build(dir)

	if !compareWithExpected(filepath.Join(dir, "test.html")) {
		t.Fail()
	}
}

func TestOnePageClean(t *testing.T) {
	dir := copyTestDir("one_page")
	defer os.RemoveAll(dir)
	site := build(dir)
	site.Clean()

	if _, err := os.Stat(filepath.Join(dir, "test.html")); os.IsExist(err) {
		t.Fail()
	}
}

func TestOneTemplate(t *testing.T) {
	dir := copyTestDir("one_template")
	defer os.RemoveAll(dir)
	build(dir)

	if !compareWithExpected(filepath.Join(dir, "test.html")) {
		t.Fail()
	}
}
