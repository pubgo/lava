package main

import (
	"flag"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/signal"
	"github.com/thejerf/suture/v4"

	_ "github.com/thejerf/suture/v4"
)

func main() {
	s := suture.NewSimple("")
	s.Add()
	s.Remove()
	s.Serve(signal.Context())

	for _, f := range flags.GetFlags() {
		assert.Exit(f.Apply(flag.CommandLine))
	}
}
