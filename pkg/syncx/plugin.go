package syncx

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
	"go.uber.org/atomic"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugin"
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
		OnFlags: func() []cli.Flag {
			return clix.Flags{
				&cli.Int64Flag{
					Name:        "concurrent",
					Usage:       "Set maximum concurrency",
					EnvVars:     typex.StrOf("lava-max-concurrency"),
					Value:       maxConcurrent,
					Destination: &maxConcurrent,
				},
			}
		},
		OnInit: func() {
			SetMaxConcurrent(maxConcurrent)
		},
		OnVars: func(w func(name string, data func() interface{})) {
			w(Name, func() interface{} {
				return typex.M{"maxConcurrent": maxConcurrent, "curConcurrent": curConcurrent.Load()}
			})
		},
	})
}
