Project=gid-http
Base=github.com/pubgo/funk
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
build:
	go build ${LDFLAGS} -v -o bin/gid *.go

vet:
	@go vet ./...
