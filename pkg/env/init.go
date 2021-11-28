package env

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	dir "github.com/mitchellh/go-homedir"
	"github.com/pubgo/xerror"
	"k8s.io/client-go/util/homedir"

	// env文件加载
	_ "github.com/joho/godotenv/autoload"

	"github.com/pubgo/lava/pkg/lavax"
)

// Prefix 系统环境变量前缀
var Prefix = os.Getenv(strings.ToUpper("env_prefix"))

// Pwd 当前目录
var Pwd = xerror.PanicStr(os.Getwd())

var Hostname = lavax.FirstNotEmpty(
	func() string { return os.Getenv("HOSTNAME") },
	func() string {
		var h, err = os.Hostname()
		xerror.Exit(err)
		return h
	},
)

func init() {
	initEnv()
}

// Home the home directory for the current user
var Home = lavax.FirstNotEmpty(
	homedir.HomeDir,
	func() string {
		var h, err = dir.Dir()
		xerror.Exit(err)
		return h
	},
)

var Namespace = lavax.FirstNotEmpty(
	func() string { return os.Getenv("NAMESPACE") },
	func() string { return os.Getenv("POD_NAMESPACE") },
	func() string {
		if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
				return ns
			}
		}
		return ""
	},
	func() string { return "default" },
)

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
