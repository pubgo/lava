package golug_version

import (
	"fmt"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xlog"
)

func init() {
	golug_trace.Log(func(_ *golug_trace.LogCtx) {
		xlog.Debug("trace [version]")
		for name, v := range List() {
			fmt.Println(name, golug_util.MarshalIndent(v))
		}
		fmt.Println()
	})
}
