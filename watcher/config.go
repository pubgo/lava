package watcher

import (
	"strings"
)

const Name = "watcher"

type Cfg struct {
	Prefix   string   `json:"prefix"`
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: "/watcher",
		Driver: "etcd",
	}
}

//  /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix string) string {
	return strings.Trim(strings.ReplaceAll(prefix, "/", "."), ".")
}
