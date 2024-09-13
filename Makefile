WORKDIR=`pwd`
Project=lava
Base=github.com/pubgo/funk
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )

LDFLAGS=-ldflags " \
-X '${Base}/version.buildTime=${BuildTime}' \
-X '${Base}/version.commitID=${CommitID}' \
-X '${Base}/version.version=${Version:-"v0.0.1-dev"}' \
-X '${Base}/version.project=${Project}' \
"

.PHONY: test
test:
	@go test -short -race -v ./... -cover

.PHONY: cover
cover:
	gocov test -tags "kcp quic" ./... | gocov-html > cover.html
	open cover.html

.PHONY: vet
vet:
	@go vet ./...

.PHONY: generate
generate:
	@go generate ./...

.PHONY: deps
deps:
	# https://github.com/protocolbuffers/protobuf
	@go install -v github.com/tinylib/msgp
	@go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

.PHONY: protobuf
protobuf:
	protobuild vendor
	protobuild gen

.PHONY: lint
lint:
	golangci-lint --version
	golangci-lint run --timeout 3m --verbose ./...

install-protoc-job:
	go install -v ./component/cloudjobs/protoc-gen-cloud-job