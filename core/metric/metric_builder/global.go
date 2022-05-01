package metric_builder

import (
	"sync/atomic"
	"unsafe"

	"github.com/uber-go/tally"
)

var g = unsafe.Pointer(&tally.NoopScope)

func GetGlobal() tally.Scope {
	return *(*tally.Scope)(atomic.LoadPointer(&g))
}
