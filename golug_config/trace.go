package golug_config

import (
	"encoding/json"
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

func init() {
	// debug and trace
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		//fmt.Println(tag)
		//fmt.Println()

		xlog.Debug("config trace")

		fmt.Println(golug_util.MarshalIndent(GetCfg().AllSettings()))
		fmt.Println()

		xlog.Debug("deps trace")
		fmt.Println(dix.Graph())
		fmt.Println()

		xlog.Debug("goroutine trace")
		data := make(map[string]interface{})
		xerror.Panic(json.Unmarshal([]byte(xprocess.Stack()), &data))
		fmt.Println(golug_util.MarshalIndent(data))
		fmt.Println()
	}))
}
