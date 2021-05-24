package env

import (
	"os"
	"strings"
)

// 环境变量的前缀
var Prefix string

func init() {
	// 获取系统默认的前缀, 环境变量前缀等
	GetWith(&Prefix, "env_prefix", "app_prefix", "service_prefix", "project_prefix")

	// 环境变量处理, key转大写, 同时把-./转换为_
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	for _, env := range os.Environ() {
		if envList := strings.SplitN(env, "=", 2); len(envList) == 2 && trim(envList[0]) != "" {
			_ = os.Unsetenv(envList[0])
			key := replacer.Replace(strings.ToUpper(trim(envList[0])))
			_ = os.Setenv(key, envList[1])
		}
	}
}
