package runenv

import (
	"os"
	"syscall"

	"github.com/pubgo/lava/pkg/env"
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
