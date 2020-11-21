Project=github.com/pubgo/golug
GOPath=$(shell go env GOPATH)
Version=$(shell git tag --sort=committerdate | tail -n 1)
GoROOT=$(shell go env GOROOT)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags " \
-X '${Project}/version.GoROOT=${GoROOT}' \
-X '${Project}/version.BuildTime=${BuildTime}' \
-X '${Project}/version.GoPath=${GOPath}' \
-X '${Project}/version.CommitID=${CommitID}' \
-X '${Project}/version.Project=${Project}' \
-X '${Project}/version.Version=${Version:-v0.0.1}' \
"

.PHONY: build
build:
	@go build ${LDFLAGS} -mod vendor -race -v -o main main.go

build_hello_test:
	go build ${LDFLAGS} -mod vendor -v -o main  example/hello/main.go

.PHONY: install
install:
	@go install ${LDFLAGS} .

.PHONY: release
release:
	@go build ${LDFLAGS} -race -v -o main main.go

.PHONY: test
test:
	@go test -race -v ./... -cover

.PHONY: proto
proto: clear gen
	protoc -I. \
   -I/usr/local/include \
   -I${GOPATH}/src \
   -I${GOPATH}/src/github.com/googleapis/googleapis \
   -I${GOPATH}/src/github.com/gogo/protobuf \
   --go_out=plugins=grpc:. \
   --golug_out=. \
	example/proto/hello/*

	protoc -I. \
   -I/usr/local/include \
   -I${GOPATH}/src \
   -I${GOPATH}/src/github.com/googleapis/googleapis \
   -I${GOPATH}/src/github.com/gogo/protobuf \
   --go_out=plugins=grpc:. \
   --golug_out=. \
	example/proto/login/*

.PHONY: clear
clear:
	rm -rf example/proto/*.go
	rm -rf example/proto/**/*.go

.PHONY: gen
gen:
	cd cmd/protoc-gen-golug && go install .

.PHONY: example
example:
	go build example/main.go