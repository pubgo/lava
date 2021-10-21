package main

import (
	_ "github.com/pubgo/lava/internal/example/services/docs"
	_ "github.com/pubgo/lava/metric/prometheus"
	_ "github.com/pubgo/lava/plugins/db/sqlite"
	_ "github.com/pubgo/lava/tracing/jaeger"
)
