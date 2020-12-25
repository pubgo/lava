package golug_env

import (
	"os"
	"strings"
)

func init() {
	// 环境变量处理, key转大写
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	for _, env := range os.Environ() {
		if _envs := strings.SplitN(env, "=", 2); len(_envs) == 2 && trim(_envs[0]) != "" {
			_ = os.Unsetenv(_envs[0])
			key := replacer.Replace(upper(trim(_envs[0])))
			_ = os.Setenv(key, _envs[1])
		}
	}
}
