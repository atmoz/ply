package fileutil

import (
	"io"
	"os"
)

func CopyFile(from, to string, options *CopyOptions) error {
	fromInfo, err := os.Stat(from)
	if err != nil {
		return err
	}

	toInfo, toErr := os.Stat(to)
	if os.IsExist(toErr) {
		hasMod := fromInfo.ModTime().After(toInfo.ModTime())
		diffSize := fromInfo.Size() != toInfo.Size()

		if !hasMod && !diffSize {
			return nil // Skip
		}
	}

	if options != nil {
		for _, regex := range options.IgnoreRegex {
			if regex.MatchString(from) {
				return nil
			}
		}
	}

	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, fromInfo.Mode())
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	return err
}
