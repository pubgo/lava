package main

import (
	"flag"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/workflow"
	"github.com/goyek/x/boot"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/buildtasks"
	"github.com/pubgo/lava/core/flags"

	_ "github.com/thejerf/suture/v4"
)

func main() {
	goyek.Undefine(workflow.PipelineAll)

	workflow.StageTest.SetDeps(append(workflow.StageTest.Deps(), buildtasks.GoLint))
	goyek.SetDefault(goyek.Define(goyek.Task{
		Name:  "all",
		Usage: "exec all tasks",
		Deps: goyek.Deps{
			workflow.StageInit,
			workflow.StageBuild,
			workflow.StageTest,
		},
	}))

	for _, f := range flags.GetFlags() {
		assert.Exit(f.Apply(flag.CommandLine))
	}
	boot.Main()
}
