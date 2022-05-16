package config

import (
	"fmt"
	"github.com/pubgo/lava/pkg/env"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/consts"
)

const resKey = "name"

func getResId(m map[string]interface{}) string {
	if m == nil {
		return consts.KeyDefault
	}

	var val, ok = m[resKey]
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
	if envPrefix == "" {
		return
	}

	var r = strings.NewReplacer("-", "_", ".", "_", "/", "_")
	envPrefix = strings.ToUpper(strings.ReplaceAll(r.Replace(strcase.ToSnake(envPrefix))+"_", "__", "_"))

	for name, val := range env.List() {
		if !strings.HasPrefix(name, envPrefix) || val == "" {
			continue
		}

		vals := strings.SplitN(val, "=", 2)
		if len(vals) != 2 {
			continue
		}

		v.Set(strings.TrimSpace(vals[0]), strings.TrimSpace(vals[1]))
	}
}
