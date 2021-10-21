package env

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	dir "github.com/mitchellh/go-homedir"
	"github.com/pubgo/xerror"
	"k8s.io/client-go/util/homedir"

	"github.com/pubgo/lava/pkg/lavax"
)

// Prefix 系统环境变量前缀
var Prefix = os.Getenv(strings.ToUpper("env_prefix"))

// Pwd 当前目录
var Pwd = xerror.PanicStr(os.Getwd())

var Hostname = lavax.FirstNotEmpty(
	func() string {
		return os.Getenv("HOSTNAME")
	}, func() string {
		var h, err = os.Hostname()
		xerror.Exit(err)
		return h
	},
)

// Home the home directory for the current user
var Home = lavax.FirstNotEmpty(
	func() string {
		var h, err = dir.Dir()
		xerror.Exit(err)
		return h
	},
	func() string {
		return homedir.HomeDir()
	},
)

func init() {
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
