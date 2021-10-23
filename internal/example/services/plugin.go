package main

import (
	_ "github.com/pubgo/lava/plugins/db/sqlite"
	_ "github.com/pubgo/lava/plugins/metric/prometheus"
	_ "github.com/pubgo/lava/plugins/tracing/jaeger"
)
