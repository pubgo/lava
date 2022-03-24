package klog

import (
	"github.com/go-logr/zapr"
	"github.com/pubgo/lava/core/logging"
	"k8s.io/klog/v2"

	"github.com/pubgo/lava/plugin"
)

// 替换klog全局log
func init() {
	plugin.RegisterProcess(
		"logging-ext-klog",
		func(p plugin.Process) {
			klog.SetLogger(zapr.NewLogger(logging.Component("klog").L()))
		})
}
