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
	IsBlock                = true
	Domain                 = "golug"
	EnvPrefix              = "golug"
	CatchSigpipe           = true
	Trace                  = false
	Port                   = 8080
	Home                   = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
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
)

// 从环境变量中获取系统默认值
func init() {
	// 获取系统默认的前缀, 环境变量前缀等
	env.GetWith(&EnvPrefix, "env_prefix", "golug_prefix")
	if EnvPrefix = trim(lower(EnvPrefix)); EnvPrefix == "" {
		EnvPrefix = "golug"
		xlog.Warnf("[env_prefix] prefix should be set, default: %s", EnvPrefix)
	}
	env.Prefix = EnvPrefix

	env.GetWith(&Domain, "domain", "golug_domain")
	if Domain = trim(lower(Domain)); Domain == "" {
		Domain = "golug"
		xlog.Warnf("[domain] prefix should be set, default: %s", Domain)
	}

	env.GetWith(&CfgType, "cfg_type")
	env.GetWith(&CfgName, "cfg_name")
	env.GetWith(&Home, "home", "cfg_dir", "config_path")
	env.GetWith(&Project, "project", "server_name")
	env.GetWith(&Mode, "mode", "run_mode")
	env.GetBoolVal(&Trace, "trace")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.IntVar(&Port, "port", Port, "the server port")
	flags.StringVarP(&Level, "level", "l", Level, "log level(debug|info|warn|error|panic|fatal)")
	flags.StringVarP(&Home, "home", "c", Home, "config home dir")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.StringVarP(&Project, "project", "p", Project, "project name")
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
