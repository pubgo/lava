package gutil

import zglob "github.com/mattn/go-zglob"

func Glob(pattern string) ([]string, error) {
	return zglob.Glob(pattern)
}
