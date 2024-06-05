package gencmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/pubgo/funk/recovery"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "gen-makefile",
		Usage: "gen makefile",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			os.Stdout.Write([]byte(`
Project=gid
Base=github.com/pubgo/lava
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)

LDFLAGS=-ldflags " \
-X '${Base}/version.buildTime=${BuildTime}' \
-X '${Base}/version.commitID=${CommitID}' \
-X '${Base}/version.version=${Version}' \
-X '${Base}/version.tag=${Tag}' \
-X '${Base}/version.project=${Project}' \
"

.PHONY: build
build-gid:
	go build ${LDFLAGS} -v -o bin/gid *.go

vet:
	@go vet ./...
`))
			return nil
		},
	}
}
