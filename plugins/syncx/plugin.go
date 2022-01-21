package syncx

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
	"go.uber.org/atomic"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "syncx"

var logs = logging.Component(Name)

// 最大goroutine数量
var maxConcurrent int64 = 100000

// 当前goroutine数量
var curConcurrent atomic.Int64

// SetMaxConcurrent 设置最大并发数
func SetMaxConcurrent(concurrent int64) {
	if runtime.NumCPU()*100 > int(concurrent) {
		panic(fmt.Sprintf("concurrent should more than %d", runtime.NumCPU()*100))
	}

	maxConcurrent = concurrent

	logs.S().Infow("set maxConcurrent", "vale", maxConcurrent)
}

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func() types.Flags {
			return types.Flags{
				&cli.Int64Flag{
					Name:        "concurrent",
					Usage:       "Set maximum concurrency",
					EnvVars:     env.KeyOf("lava-max-concurrency"),
					Value:       maxConcurrent,
					Destination: &maxConcurrent,
				},
			}
		},
		OnInit: func(p plugin.Process) { SetMaxConcurrent(maxConcurrent) },
		OnWatch: func(name string, r *types.WatchResp) error {
			var concurrent int64
			logutil.LogOrPanic(logs.L(), "max concurrent decode", func() error { return r.Decode(&concurrent) })
			SetMaxConcurrent(concurrent)
			return nil
		},
		OnVars: func(v types.Vars) {
			v.Do(Name, func() interface{} {
				return types.M{"maxConcurrent": maxConcurrent, "curConcurrent": curConcurrent.Load()}
			})
		},
	})
}
