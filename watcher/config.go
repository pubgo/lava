package watcher

import (
	"strings"
)

var Name = "watcher"
var cfgList []Cfg

type Cfg struct {
	Project string `json:"project"`
	Driver  string `json:"driver"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Project: "hello",
		Driver:  "etcdv3",
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
