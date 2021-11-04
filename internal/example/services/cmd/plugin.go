package main

import (
	_ "github.com/pubgo/lava/clients/db/sqlite"
	_ "github.com/pubgo/lava/plugins/metric/prometheus"
	_ "github.com/pubgo/lava/plugins/tracing/jaeger"
)
