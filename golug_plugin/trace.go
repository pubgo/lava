package golug_plugin

import (
	"fmt"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/xlog"
)

func init() {
	golug_trace.Log(func(_ *golug_trace.LogCtx) {
		xlog.Debug("trace [plugin]")
		fmt.Println(String())
		fmt.Println()
	})
}
