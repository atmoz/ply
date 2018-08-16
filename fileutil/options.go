package fileutil

import "regexp"

type CopyOptions struct {
	IgnoreRegex []*regexp.Regexp
}
