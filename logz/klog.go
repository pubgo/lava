package logz

import (
	"github.com/go-logr/zapr"
	"k8s.io/klog/v2"
)

// 替换klog全局log
func init() {
	On(func(*Log) {
		klog.SetLogger(zapr.NewLogger(getName("klog")))
	})
}
