package runenv

import (
	"os"
	"syscall"
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
