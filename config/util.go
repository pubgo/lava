package config

import (
	"fmt"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/xerror"
	"os"
	"path/filepath"
)

const pkgKey = "name"

func getPkgId(m map[string]interface{}) string {
	if m == nil {
		return consts.KeyDefault
	}

	var val, ok = m[pkgKey]
	if !ok || val == nil {
		return consts.KeyDefault
	}

	return fmt.Sprintf("%v", val)
}

// getPathList 递归得到当前目录到跟目录中所有的目录路径
//	[./, ../, ../../, ..., /]
func getPathList() (paths []string) {
	var wd = xerror.PanicStr(filepath.Abs(""))
	for {
		if len(wd) == 0 || os.IsPathSeparator(wd[len(wd)-1]) {
			break
		}

		paths = append(paths, wd)
		wd = filepath.Dir(wd)
	}
	return
}

func strMap(strList []string, fn func(str string) string) []string {
	for i := range strList {
		strList[i] = fn(strList[i])
	}
	return strList
}

func Decode[Cfg any](c Config, name string) map[string]Cfg {
	var cfgMap = make(map[string]Cfg)
	xerror.PanicF(c.Decode(name, &cfgMap), "config decode failed, name=%s", name)
	return cfgMap
}

func MakeClient[Cfg any, Client any](c Config, name string, callback func(key string, cfg Cfg) Client) map[string]Client {
	var cfgMap = Decode[Cfg](c, name)
	var clientMap = make(map[string]Client)
	for key := range cfgMap {
		clientMap[key] = callback(key, cfgMap[key])
	}
	return clientMap
}
