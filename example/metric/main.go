package main

import "github.com/pubgo/lug/metric"

func main() {
	metric.CreateSummary(metric.SummaryOpts{})
	metric.Summary("hello", 11, nil)
}
