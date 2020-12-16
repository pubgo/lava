package golug_watcher

import (
	"fmt"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xlog"
)

func init() {
	// debug and trace
	golug_trace.Log(func(_ *golug_trace.LogCtx) {
		xlog.Debug("trace [log] config")
		var dt []string
		dataCallback.Range(func(key, _ interface{}) bool { dt = append(dt, key.(string)); return true })
		fmt.Println(golug_util.MarshalIndent(dt))
		fmt.Println()
	})
}
