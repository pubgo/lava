WORKDIR=`pwd`
Domain=lava
VersionBase=github.com/pubgo/lava
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )

LDFLAGS=-ldflags " \
-X '${VersionBase}/version.BuildTime=${BuildTime}' \
-X '${VersionBase}/version.CommitID=${CommitID}' \
-X '${VersionBase}/version.Version=${Version}' \
-X '${VersionBase}/version.Tag=${Tag}' \
-X '${VersionBase}/version.Domain=${Domain}' \
-X '${VersionBase}/version.Data=hello' \
"

.PHONY: build
build:
	go build ${LDFLAGS} -v -o main cmd/*.go

.PHONY: run
run:
	go run ${LDFLAGS} -v cmd/*.go v

.PHONY: test./ma
test:
	@go test -short -race -v ./... -cover

vet:
	@go vet ./...

generate:
	@go generate ./...

lint:
	@golangci-lint run --skip-dirs-use-default --timeout 3m0s
