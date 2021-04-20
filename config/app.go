package config

import (
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// 默认的全局配置
var (
	CfgType      = "yaml"
	CfgName      = "config"
	IsBlock      = true
	Domain       = "lug"
	CatchSigpipe = true
	Trace        = false
	Port         = 8080
	DebugPort    = 8088
	Home         = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
	CfgPath      = ""
	Project      = "lug"
	Level        = "debug"
	Mode         = Dev.String()
	Signal       = syscall.Signal(0)
	trim         = strings.TrimSpace
	lower        = strings.ToLower
	cfg          = &Config{Viper: viper.New()}
)

const (
	Dev RunMode = iota + 1
	Test
	Stag
	Prod
	Release
)

// RunMode 项目运行模式
type RunMode uint8

func (t RunMode) String() string {
	switch t {
	case 1:
		return "dev"
	case 2:
		return "test"
	case 3:
		return "stag"
	case 4:
		return "prod"
	case 5:
		return "release"
	default:
		xlog.Errorf("running mode(%d) not match", t)
		return consts.Unknown
	}
}

// 从环境变量中获取系统默认值
func init() {
	// 获取系统默认的前缀, 环境变量前缀等
	env.GetWith(&env.Prefix, "env_prefix", "lug_prefix", "app_prefix", "service_prefix", "project_prefix")

	// 服务group
	env.GetWith(&Domain, "domain", "app_domain", "project_domain")
	if Domain = trim(lower(Domain)); Domain == "" {
		Domain = "lug"
		xlog.Warnf("[domain] is null, set default: %s", Domain)
	}

	env.GetWith(&CfgType, "cfg_type", "config_type")
	env.GetWith(&CfgName, "cfg_name", "config_name")
	env.GetWith(&Home, "project_home", "config_home", "cfg_dir", "config_path", "config_dir")
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
	return Mode == Dev.String()
}

func IsTest() bool {
	return Mode == Test.String()
}

func IsStag() bool {
	return Mode == Stag.String()
}

func IsProd() bool {
	return Mode == Prod.String()
}

func IsRelease() bool {
	return Mode == Release.String()
}
