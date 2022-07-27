WORKDIR=`pwd`
Domain=lava
VersionBase=github.com/pubgo/lava
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )

LDFLAGS=-ldflags " \
-X '${Base}/version.BuildTime=${BuildTime}' \
-X '${Base}/version.CommitID=${CommitID}' \
-X '${Base}/version.Version=${Version}' \
-X '${Base}/version.Tag=${Tag}' \
-X '${Base}/version.Domain=${Domain}' \
-X '${Base}/version.Data=hello' \
"

.PHONY: build
build:
	go build ${LDFLAGS} -v -o bin/main *.go

vet:
	@go vet ./...
