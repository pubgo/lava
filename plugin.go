package lava

import _ "github.com/pubgo/lava/plugins"

// 加载插件
import (
	// set GOMAXPROCS
	_ "github.com/pubgo/lava/internal/modules/automaxprocs"

	// gc plugin
	_ "github.com/pubgo/lava/internal/modules/gcnotifier"

	// metric
	//_ "github.com/pubgo/lava/core/metric/metric_builder"

	// 用于系统诊断
	_ "github.com/pubgo/lava/internal/modules/gops"
)

// 加载middleware, 注意加载顺序
import (
	_ "github.com/pubgo/lava/logging/middleware"

	_ "github.com/pubgo/lava/core/requestid"
)
