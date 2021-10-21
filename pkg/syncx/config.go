package syncx

import (
	"fmt"
	"runtime"

	"go.uber.org/atomic"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

const Name = "goroutine"

var maxConcurrent uint32 = 100000
var curConcurrent atomic.Uint32

func SetMaxConcurrent(concurrent uint32) {
	if runtime.NumCPU()*100 > int(concurrent) {
		panic(fmt.Sprintf("concurrent should more than %d", runtime.NumCPU()*100))
	}

	maxConcurrent = concurrent

	logz.Named(Name).Infof("set maxConcurrent=>%d", maxConcurrent)
}

func init() {
	vars.Watch(Name, func() interface{} {
		return typex.M{"maxConcurrent": maxConcurrent, "curConcurrent": curConcurrent.Load()}
	})
}
