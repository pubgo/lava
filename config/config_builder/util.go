package config_builder

import (
	"os"
	"path/filepath"

	"github.com/pubgo/xerror"
)

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
