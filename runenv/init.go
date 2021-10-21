package runenv

import (
	"os"
	"syscall"

	"github.com/pubgo/lava/pkg/env"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	Domain                 = "lava"
	CatchSigpipe           = false
	Block                  = true
	Trace                  = false
	Addr                   = ":8080"
	DebugAddr              = ":8081"
	Project                = "lava"
	Level                  = "debug"
	Mode                   = "dev"
	Signal       os.Signal = syscall.Signal(0)
)

// 从环境变量中获取系统默认值
func init() {
	// 运行环境
	env.GetWith(&Mode, "run_mode", "run_env")

	// 服务group,domain
	env.GetWith(&Domain, "domain", "service_domain", "project_domain")

	// 开启trace
	env.GetBoolVal(&Trace, "trace", "trace_log", "tracelog")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&Addr, "addr", "a", Addr, "server(http|grpc|ws|...) address")
	flags.StringVar(&DebugAddr, "debug-addr", DebugAddr, "debug server address")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.StringVarP(&Level, "level", "l", Level, "log level(debug|info|warn|error|panic|fatal)")
	flags.BoolVar(&CatchSigpipe, "catch-sigpipe", CatchSigpipe, "catch and ignore SIGPIPE on stdout and stderr if specified")
	return flags
}
