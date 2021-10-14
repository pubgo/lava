package syncx

import (
	"fmt"
	"runtime"

	"github.com/pubgo/lava/logz"
	"go.uber.org/atomic"
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
	logz.Named(Name).Infof("default maxConcurrent=>%d", maxConcurrent)
}

//调度