package config

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pubgo/golug/env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	CfgType                = "yaml"
	CfgName                = "config"
	IsBlock                = true
	Domain                 = "golug"
	CatchSigpipe           = true
	Trace                  = false
	Port                   = 8080
	DebugPort              = 8088
	Home                   = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
	CfgPath                = ""
	Project                = "golug"
	Level                  = "debug"
	Mode                   = RunMode.Dev
	Signal       os.Signal = syscall.Signal(0)
	// RunEnvMode 项目运行模式
	RunMode = struct {
		Dev     string
		Test    string
		Stag    string
		Prod    string
		Release string
	}{Dev: "dev", Test: "test", Stag: "stag", Prod: "prod", Release: "release"}
	trim  = strings.TrimSpace
	lower = strings.ToLower
	cfg   *Config
)

// 从环境变量中获取系统默认值
func init() {
	// 获取系统默认的前缀, 环境变量前缀等
	env.GetWith(&env.Prefix, "env_prefix", "golug_prefix", "app_prefix", "service_prefix", "project_prefix")

	// 服务group
	env.GetWith(&Domain, "domain", "app_domain", "project_domain")
	if Domain = trim(lower(Domain)); Domain == "" {
		Domain = "golug"
		xlog.Warnf("[domain] is null, set default: %s", Domain)
	}

	env.GetWith(&CfgType, "cfg_type", "config_type")
	env.GetWith(&CfgName, "cfg_name", "config_name")
	env.GetWith(&Home, "home", "cfg_dir", "config_path", "config_dir")
	env.GetWith(&Mode, "mode", "run_mode", "run_env", "project_mode")
	env.GetBoolVal(&Trace, "trace", "trace_log")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.IntVarP(&Port, "port", "p", Port, "the application server port")
	flags.IntVar(&DebugPort, "dp", DebugPort, "the debug server port")
	flags.StringVarP(&Level, "level", "l", Level, "log level(debug|info|warn|error|panic|fatal)")
	flags.StringVarP(&CfgPath, "cfg", "c", CfgPath, "config path")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.BoolVarP(&IsBlock, "block", "b", IsBlock, "enable signal block")
	flags.BoolVar(&CatchSigpipe, "catch-sigpipe", CatchSigpipe, "catch and ignore SIGPIPE on stdout and stderr if specified")
	return flags
}

func IsDev() bool {
	return Mode == RunMode.Dev
}

func IsTest() bool {
	return Mode == RunMode.Test
}

func IsStag() bool {
	return Mode == RunMode.Stag
}

func IsProd() bool {
	return Mode == RunMode.Prod
}

func IsRelease() bool {
	return Mode == RunMode.Release
}
