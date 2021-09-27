WORKDIR=`pwd`
Domain=lugo
VersionBase=github.com/pubgo/lug/example/gid
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )
TAG=$(shell git describe --abbrev=0 --tags)
LDFLAGS=-ldflags " \
-X '${VersionBase}/version.BuildTime=${BuildTime}' \
-X '${VersionBase}/version.CommitID=${CommitID}' \
-X '${VersionBase}/version.Version=${Version}' \
-X '${VersionBase}/version.Domain=${Domain}' \
-X '${VersionBase}/version.Data=hello' \
"

.PHONY: build
build:
	@go build ${LDFLAGS} -mod vendor -v -o main ./*.go

.PHONY: test./ma
test:
	@go test -short -race -v ./... -cover

ci:
	@golangci-lint run -v --timeout=5m


proto-gen:
	rm -rf proto/*.go
	rm -rf proto/*.json
	flerken protoc ls
	flerken protoc gen

proto-vendor:
	rm -rf proto/hello/*.go
	rm -rf proto/hello/*.json
	flerken protoc vendor-rm
	flerken protoc vendor
	flerken protoc ls
	flerken protoc gen

cover:
	gocov test -tags "kcp quic" ./... | gocov-html > cover.html
	open cover.html

vet:
	@go vet ./...

generate:
	@go generate ./...

.PHONY: install
install:
	@go install -v github.com/tinylib/msgp
	@go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

lint:
	@golangci-lint run --skip-dirs-use-default --timeout 3m0s
