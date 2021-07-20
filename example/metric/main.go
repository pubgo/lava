package main

import "github.com/pubgo/lug/metric"

func main() {
	metric.Summary("hello", 11, nil)
}
