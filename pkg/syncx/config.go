package syncx

import (
	"fmt"
	"runtime"

	"go.uber.org/atomic"

	"github.com/pubgo/lava/logger"
)

const Name = "goroutine"

var maxConcurrent uint32 = 100000
var curConcurrent atomic.Uint32

func SetMaxConcurrent(concurrent uint32) {
	if runtime.NumCPU()*100 > int(concurrent) {
		panic(fmt.Sprintf("[concurrent] should more than %d", runtime.NumCPU()*100))
	}

	maxConcurrent = concurrent

	logger.GetSugar(Name).Infof("current set maxConcurrent=>%d", maxConcurrent)
}

func init() {
	logger.GetSugar(Name).Infof("default maxConcurrent=>%d", maxConcurrent)
}
