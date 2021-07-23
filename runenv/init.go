package runenv

import (
	"os"
	"syscall"

	"github.com/pubgo/lug/pkg/env"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	Domain                 = "lug"
	CatchSigpipe           = true
	Block                  = true
	Trace                  = false
	Addr                   = ":8080"
	Project                = "lug"
	Level                  = "debug"
	Mode                   = "dev"
	Signal       os.Signal = syscall.Signal(0)
)

// 从环境变量中获取系统默认值
func init() {
	// 运行环境
	env.GetWith(&Mode, "mode", "runenv", "run_mode", "run_env")

	// 服务group,domain
	env.GetWith(&Domain, "domain", "service_domain", "project_domain")

	// 开启trace
	env.GetBoolVal(&Trace, "trace", "trace_log")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.StringVarP(&Addr, "addr", "a", Addr, "service address")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.StringVarP(&Level, "level", "l", Level, "log level(debug|info|warn|error|panic|fatal)")
	flags.BoolVar(&CatchSigpipe, "catch-sigpipe", CatchSigpipe, "catch and ignore SIGPIPE on stdout and stderr if specified")
	return flags
}
