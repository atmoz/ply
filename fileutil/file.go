package fileutil

import (
	"io"
	"os"
	"path/filepath"
)

func CopyFile(from, to string, options *CopyOptions) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	fromInfo, err := os.Stat(from)
	if err != nil {
		return err
	}

	if options != nil && options.IgnoreRegex != nil {
		if options.IgnoreRegex.MatchString(filepath.Base(from)) {
			return nil
		}
	}

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, fromInfo.Mode())
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	return err
}
