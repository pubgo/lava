package main

import (
	"context"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/internal/example/grpc/internal/bootstrap"
)

func main() {
	log.SetEnableChecker(func(ctx context.Context, lvl log.Level, nameOrMessage string, fields log.Map) bool {
		if nameOrMessage == "eval type value" {
			return false
		}

		if nameOrMessage == "grpc-server" {
			return false
		}

		return true
	})

	bootstrap.Main()
}
