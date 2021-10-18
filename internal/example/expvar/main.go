package main

import (
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"

	"net/http"
)

func main() {
	exp.Exp(metrics.DefaultRegistry)
	http.ListenAndServe(":8081", nil)
}
