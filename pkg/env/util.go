package env

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

// initEnv
// 环境变量处理, key转大写, 同时把-./转换为_
// a-b=>a_b, a.b=>a_b, a/b=>a_b, HelloWorld=>hello_world
func init() {
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	for _, env := range os.Environ() {
		if envs := strings.SplitN(env, "=", 2); len(envs) == 2 && trim(envs[0]) != "" {
			_ = os.Unsetenv(envs[0])
			key := replacer.Replace(strcase.ToSnake(trim(envs[0])))
			_ = os.Setenv(strings.ToUpper(key), envs[1])
		}
	}
}
