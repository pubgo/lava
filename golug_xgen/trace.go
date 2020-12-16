package golug_xgen

import (
	"fmt"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/xlog"
)

func init() {
	golug_trace.Log(func(_ *golug_trace.LogCtx) {
		xlog.Debug("trace [data]")
		for k, v := range List() {
			fmt.Printf("%#v: \n\t%#v\n\n", k, v)
		}
		fmt.Println()
	})
}
