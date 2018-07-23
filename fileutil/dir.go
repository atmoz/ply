package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CopyDirectory(fromDir, toDir string) error {
	fromInfo, err := os.Stat(fromDir)
	if err != nil {
		return err
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
			err = CopyDirectory(fromFile, toFile)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(fromFile, toFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
