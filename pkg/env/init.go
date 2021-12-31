package env

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

func init() {
	// 加载.env
	Load()

	initEnv()
}

func initEnv() {
	// 环境变量处理, key转大写, 同时把-./转换为_
	// a-b=>a_b, a.b=>a_b, a/b=>a_b, HelloWorld=>hello_world
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	for _, env := range os.Environ() {
		if envList := strings.SplitN(env, "=", 2); len(envList) == 2 && trim(envList[0]) != "" {
			_ = os.Unsetenv(envList[0])
			key := replacer.Replace(strcase.ToSnake(trim(envList[0])))
			_ = os.Setenv(strings.ToUpper(key), envList[1])
		}
	}
}
