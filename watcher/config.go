package watcher

import (
	"strings"
)

var Name = "watcher"

type Cfg struct {
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "etcdv3",
	}
}

// KeyWithDot [abc,ddd/ss,a,.c] --> abc.ddd/ss.a.c
func KeyWithDot(key ...string) string {
	return strings.ReplaceAll(strings.Join(key, "."), "..", ".")
}

//  /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix string) string {
	return strings.Trim(strings.ReplaceAll(prefix, "/", "."), ".")
}
