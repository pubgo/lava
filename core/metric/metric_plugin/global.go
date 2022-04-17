package metric_plugin

import (
	"sync/atomic"
	"unsafe"

	"github.com/uber-go/tally"

	"github.com/pubgo/lava/inject"
)

var g = unsafe.Pointer(&tally.NoopScope)

func GetGlobal() tally.Scope {
	return *(*tally.Scope)(atomic.LoadPointer(&g))
}

func init() {
	// 注入依赖scope
	inject.Register((*tally.Scope)(nil), func(obj inject.objectImpl, field inject.fieldImpl) (interface{}, bool) {
		return GetGlobal(), true
	})
}
