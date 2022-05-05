package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/consts"
)

const _resIdKey = "name"

func getResId(m map[string]interface{}) string {
	if m == nil {
		return consts.KeyDefault
	}

	var val, ok = m[_resIdKey]
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

func loadEnv(envPrefix string, v *viper.Viper) {
	var r = strings.NewReplacer("-", "_", ".", "_", "__", "_", "/", "_")
	envPrefix = strings.ReplaceAll(r.Replace(envPrefix)+"_", "__", "_")

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, envPrefix) {
			continue
		}

		env = strings.TrimPrefix(env, envPrefix)
		var envs = strings.SplitN(env, "=", 2)
		if len(envs) != 2 {
			continue
		}

		if strings.TrimSpace(envs[1]) == "" {
			continue
		}

		envs = strings.SplitN(envs[1], "=", 2)
		if len(envs) != 2 {
			continue
		}

		v.Set(strings.TrimSpace(envs[0]), strings.TrimSpace(envs[1]))
	}
}
