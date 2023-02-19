WORKDIR=`pwd`
Domain=lava
Project=test-grpc
Base=github.com/pubgo/funk
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )

LDFLAGS=-ldflags " \
-X '${Base}/version.buildTime=${BuildTime}' \
-X '${Base}/version.commitID=${CommitID}' \
-X '${Base}/version.version=${Version:=v0.0.1-dev}' \
-X '${Base}/version.tag=${Tag}' \
-X '${Base}/version.domain=${Domain}' \
-X '${Base}/version.project=${Project}' \
-X '${Base}/version.data=hello' \
"

.PHONY: build
build:
	go build ${LDFLAGS} -mod vendor -v -o main cmd/lava/main.go

.PHONY: install
install:
	@go build ${LDFLAGS} -mod vendor -tags trace -o lava -v cmd/lava/*.go
	@mv lava ${GOPATH}/bin/lava

build_hello_test:
	@go build ${LDFLAGS} -mod vendor -v -o main  example/hello/main.go

.PHONY: test./ma
test:
	@go test -short -race -v ./... -cover

ci:
	@golangci-lint run -v --timeout=5m

.PHONY: gen
proto-plugin-gen:
	cd cmd/protoc-gen-errors && go install -v .

.PHONY: example
example:
	go build ${LDFLAGS} -mod vendor -v -o main example/*.go

docker:
	docker build -t lug .

build-all:
	go build -tags "kcp quic" ./...

cover:
	gocov test -tags "kcp quic" ./... | gocov-html > cover.html
	open cover.html

check-libs:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go list -m -u all | column -t

update-libs:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go get -u -v ./...

mod-tidy:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go mod tidy

vet:
	@go vet ./...

generate:
	@go generate ./...

.PHONY: deps
deps:
	# https://github.com/protocolbuffers/protobuf
	@go install -v github.com/tinylib/msgp
	@go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

lint:
	@golangci-lint run --skip-dirs-use-default --timeout 3m0s
