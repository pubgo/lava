package runenv

import (
	"os"
	"syscall"

	"github.com/pubgo/lava/version"
)

// 默认的全局配置
var (
	Domain                 = version.Domain
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
