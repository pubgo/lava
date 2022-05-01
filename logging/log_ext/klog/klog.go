package klog

import (
	"github.com/go-logr/zapr"
	"k8s.io/klog/v2"

	"github.com/pubgo/lava/logging"
)

// 替换klog全局log
func init() {
	klog.SetLogger(zapr.NewLogger(logging.Component("klog").L()))
}
