package syncx

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
	"go.uber.org/atomic"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "syncx"

var maxConcurrent int64 = 100000
var curConcurrent atomic.Int64
var logs = logz.New(Name)

// SetMaxConcurrent 设置最大并发数
func SetMaxConcurrent(concurrent int64) {
	if runtime.NumCPU()*100 > int(concurrent) {
		panic(fmt.Sprintf("concurrent should more than %d", runtime.NumCPU()*100))
	}

	maxConcurrent = concurrent

	logs.Infof("set maxConcurrent=>%d", maxConcurrent)
}

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func() types.Flags {
			return types.Flags{
				&cli.Int64Flag{
					Name:        "concurrent",
					Usage:       "Set maximum concurrency",
					EnvVars:     types.EnvOf("lava-max-concurrency"),
					Value:       maxConcurrent,
					Destination: &maxConcurrent,
				},
			}
		},
		OnInit: func(p plugin.Process) {
			SetMaxConcurrent(maxConcurrent)
		},
		OnVars: func(v types.Vars) {
			v.Do(Name, func() interface{} {
				return types.M{"maxConcurrent": maxConcurrent, "curConcurrent": curConcurrent.Load()}
			})
		},
	})
}
