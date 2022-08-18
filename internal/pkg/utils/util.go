package utils

import (
	"os"
	"strings"
	"time"

	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/consts"
)

func DotJoin(str ...string) string {
	return strings.Join(str, ".")
}

// DirExists function to check if directory exists?
func DirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		// path is a directory
		return true
	}
	return false
}

// FileExists function to check if file exists?
func FileExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}

func FirstFnNotEmpty(fx ...func() string) string {
	for i := range fx {
		if s := fx[i](); s != "" {
			return s
		}
	}
	return ""
}

func FirstNotEmpty(strs ...string) string {
	for i := range strs {
		if s := strs[i]; s != "" {
			return s
		}
	}
	return ""
}

func IfEmpty(str string, fx func()) {
	if str == "" {
		fx()
	}
}

func GetDefault(names ...string) string {
	var name = consts.KeyDefault
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	return name
}

func Cost(fn func()) (dur time.Duration, err error) {
	defer func(t time.Time) { dur = time.Since(t) }(time.Now())
	defer recovery.Err(&err)
	fn()
	return
}
