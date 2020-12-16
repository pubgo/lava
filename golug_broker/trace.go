package golug_broker

import (
	"fmt"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xlog"
)

func init() {
	golug_trace.Log(func(_ *golug_trace.LogCtx) {
		xlog.Debug("trace [broker] list")
		fmt.Println(golug_util.MarshalIndent(List()))
		fmt.Println()
	})
}
