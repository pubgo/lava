package pathutil

import glob "github.com/mattn/go-zglob"

func Glob(pattern string) ([]string, error) {
	return glob.Glob(pattern)
}
