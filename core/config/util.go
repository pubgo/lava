package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/merge"
)

func getComponentName(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return defaultComponentKey
	}

	var val, ok = m[componentConfigKey]
	if !ok || val == nil {
		return defaultComponentKey
	}

	return fmt.Sprintf("%v", val)
}

// getPathList 递归得到当前目录到跟目录中所有的目录路径
//
//	paths: [./, ../, ../../, ..., /]
func getPathList() (paths []string) {
	var wd = assert.Must1(filepath.Abs(""))
	for len(wd) > 0 && !os.IsPathSeparator(wd[len(wd)-1]) {
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

func getCfgData() interface{} {
	var cfg = New()
	return map[string]any{
		"cfg_type":   defaultConfigType,
		"cfg_name":   defaultConfigName,
		"home":       CfgDir,
		"cfg_path":   CfgPath,
		"all_key":    cfg.AllKeys(),
		"all_config": cfg.All(),
	}
}

func Load[T any]() T {
	var c = New()
	var cfg T
	assert.Must(c.Unmarshal(&cfg))
	return cfg
}

func Merge[A any, B any](dst A, src *B) *A {
	return merge.Copy(generic.Ptr(dst), src).Unwrap()
}
