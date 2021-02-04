package golug_watcher

import (
	"strings"

	"github.com/pubgo/golug/golug_config"
)

var Name = "watcher"

type Cfg struct {
	Project string `json:"project"`
	Driver  string `json:"driver"`
	Name    string `json:"name"`
}

func GetCfg() (cfg map[string]Cfg) {
	golug_config.Decode(Name, &cfg)
	return
}

func GetDefaultCfg() Cfg {
	return Cfg{}
}

// KeyWithDot [abc,ddd/ss,a,.c] --> abc.ddd/ss.a.c
func KeyWithDot(key ...string) string {
	return strings.ReplaceAll(strings.Join(key, "."), "..", ".")
}

//  /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix string) string {
	return strings.Trim(strings.ReplaceAll(prefix, "/", "."), ".")
}
