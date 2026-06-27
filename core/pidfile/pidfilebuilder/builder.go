// Package pidfilebuilder 提供 lifecycle 钩子，在服务启动后写入 PID 文件、停止后删除。
package pidfilebuilder

import (
	"context"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/pidfile"
)

// New 返回 lifecycle.Handler：AfterStart 写入 PID，AfterStop 删除 PID 文件。
func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		lc.AfterStart(func(ctx context.Context) error { return pidfile.Save().GetErr() })
		lc.AfterStop(func(ctx context.Context) error { return pidfile.Remove().GetErr() })
	}
}
