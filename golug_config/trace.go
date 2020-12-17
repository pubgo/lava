package golug_config

import (
	"encoding/json"
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/internal/golug_util"
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

		//expvar.Publish(Name+"1", expvar.Func(func() interface{} { return GetCfg().AllKeys() }))
		//expvar.Publish(Name+"4", expvar.Func(func() interface{} { return golug_util.MarshalIndent(GetCfg().AllSettings()) }))
		//expvar.NewString(Name + "2").Set(dix.Graph())
		//expvar.NewString(Name + "3").Set(xprocess.Stack())

		xlog.Debug("trace [config]")
		fmt.Println(golug_util.MarshalIndent(GetCfg().AllSettings()))
		fmt.Println()

		xlog.Debug("trace [deps]")
		fmt.Println(dix.Graph())
		fmt.Println()

		xlog.Debug("trace [goroutine]")
		data := make(map[string]interface{})
		xerror.Panic(json.Unmarshal([]byte(xprocess.Stack()), &data))
		fmt.Println(golug_util.MarshalIndent(data))
		fmt.Println()
	}))
}
