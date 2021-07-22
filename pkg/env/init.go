package env

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

// Prefix 环境变量的前缀
var Prefix string

func init() {
	// env_prefix 获取系统环境变量前缀
	envDt, ok := Lookup("env_prefix")
	envDt = trim(envDt)
	if ok && envDt != "" {
		Prefix = envDt
	}

	// 环境变量处理, key转大写, 同时把-./转换为_
	// a-b=>a_b, a.b=>a_b, a/b=>a_b, HelloWorld=>hello_world
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	for _, env := range os.Environ() {
		if envList := strings.SplitN(env, "=", 2); len(envList) == 2 && trim(envList[0]) != "" {
			_ = os.Unsetenv(envList[0])
			key := replacer.Replace(strcase.ToSnake(trim(envList[0])))
			_ = os.Setenv(strings.ToLower(key), envList[1])
		}
	}
}
