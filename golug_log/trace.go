package golug_log

import (
	"fmt"

	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
)

func trace(cfg xlog_config.Config) {
	if !golug_config.Trace {
		return
	}

	xlog.Debug("log trace")
	fmt.Println(golug_util.MarshalIndent(cfg))
	fmt.Println()
}
