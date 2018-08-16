package fileutil

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func AbsRootLimit(rootPath, path string) (string, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}

	var absPath string
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		absPath = filepath.Join(absRoot, path)
	}

	if strings.HasPrefix(absPath, absRoot) {
		return absPath, nil
	} else {
		return "", errors.New(path + " is not a subdirectory of " + absRoot)
	}
}

func CopyDirectory(fromDir, toDir string, options *CopyOptions) error {
	fromInfo, err := os.Stat(fromDir)
	if err != nil {
		return err
	}

	if options != nil {
		for _, regex := range options.IgnoreRegex {
			if regex.MatchString(fromDir) {
				return nil
			}
		}
	}

	err = os.MkdirAll(toDir, fromInfo.Mode())
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(fromDir)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		fromFile := filepath.Join(fromDir, fileInfo.Name())
		toFile := filepath.Join(toDir, fileInfo.Name())

		if fileInfo.IsDir() {
			err = CopyDirectory(fromFile, toFile, options)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(fromFile, toFile, options)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
