package buildtasks

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var GoLint = goyek.Define(goyek.Task{
	Name:  "go-lint",
	Usage: "golangci-lint run --fix",
	Action: func(a *goyek.A) {
		if !cmd.Exec(a, "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest") {
			return
		}
		cmd.Exec(a, "golangci-lint run --fix")
	},
})
