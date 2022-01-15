package klog

import (
	"github.com/go-logr/zapr"
	"k8s.io/klog/v2"

	"github.com/pubgo/lava/logger"
)

// 替换klog全局log
func init() {
	logger.On(func(*logger.Event) {
		klog.SetLogger(zapr.NewLogger(logger.Component("klog").L()))
	})
}
