package runenv

import (
	"os"
	"strings"
	"syscall"

	"github.com/pubgo/lug/pkg/env"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	Domain                 = "lug"
	CatchSigpipe           = true
	Trace                  = false
	Addr                   = ":8080"
	Project                = "lug"
	Level                  = "debug"
	Mode                   = "dev"
	Signal       os.Signal = syscall.Signal(0)
)

// 从环境变量中获取系统默认值
func init() {
	// 获取系统默认的前缀, 环境变量前缀等
	env.GetWith(&env.Prefix, "env_prefix", "lug_prefix", "app_prefix", "service_prefix", "project_prefix")

	// 服务group domain
	env.GetWith(&Domain, "domain", "app_domain", "project_domain")
	if Domain = strings.TrimSpace(strings.ToLower(Domain)); Domain == "" {
		Domain = "lug"
	}

	env.GetWith(&Mode, "mode", "run_mode", "run_env", "project_mode")
	env.GetBoolVal(&Trace, "trace", "trace_log")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&Addr, "addr", "a", Addr, "the application server addr")
	flags.StringVarP(&Level, "level", "l", Level, "log level(debug|info|warn|error|panic|fatal)")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.BoolVar(&CatchSigpipe, "catch-sigpipe", CatchSigpipe, "catch and ignore SIGPIPE on stdout and stderr if specified")
	return flags
}
