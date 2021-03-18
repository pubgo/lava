package watcher

import (
	"strings"
)

var Name = "watcher"
var Prefix = "/watchers"

type Cfg struct {
	Prefix   string   `json:"prefix"`
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: Prefix,
		Driver: "etcdv3",
	}
}

//  /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix string) string {
	return strings.Trim(strings.ReplaceAll(prefix, "/", "."), ".")
}
