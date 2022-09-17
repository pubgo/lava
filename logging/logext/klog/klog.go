package klog

import (
	"github.com/go-logr/zapr"
	"github.com/pubgo/dix/di"
	"k8s.io/klog/v2"

	"github.com/pubgo/lava/logging"
)

// 替换klog全局log
func init() {
	di.Provide(func() logging.ExtLog {
		return func(logger *logging.Logger) {
			klog.SetLogger(zapr.NewLogger(logging.ModuleLog(logger, "klog").L()))
		}
	})
}
