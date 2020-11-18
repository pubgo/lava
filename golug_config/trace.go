package golug_config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

func init() {
	// debug and trace
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !Trace {
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

		xlog.Debug("env trace")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, Domain) {
				fmt.Println(env)
			}
		}
		fmt.Println()
	}))

	// 运行环境检查
	xerror.Panic(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		var m = RunMode
		switch Mode {
		case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
		default:
			xerror.Panic(xerror.Fmt("running mode does not match, mode: %s", Mode))
		}

		// 判断debug模式
		switch Mode {
		case RunMode.Dev, RunMode.Test, "":
			Debug = true
		}
	}))
}
