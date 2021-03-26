WORKDIR=`pwd`
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
-X '${Project}/version.Version=${Version}' \
"

default: run

.PHONY: build
build:
	@go build ${LDFLAGS} -mod vendor -v -o main cmds/golug/main.go

build_hello_test:
	@go build ${LDFLAGS} -mod vendor -v -o main  example/hello/main.go

.PHONY: test
test:
	@go test -short -race -v ./... -cover

.PHONY: proto
proto: clear gen
	protoc -I. \
   -I/usr/local/include \
   -I${GOPATH}/src \
   -I${GOPATH}/src/github.com/googleapis/googleapis \
   -I${GOPATH}/src/github.com/gogo/protobuf \
   --go_out=plugins=grpc:. \
   --go_opt=paths=source_relative \
   --grpc-gateway_out=. \
   --grpc-gateway_opt=paths=source_relative \
   --grpc-gateway_opt=logtostderr=true \
   --golug_out=. \
	example/proto/hello/*.proto

	protoc -I. \
   -I/usr/local/include \
   -I${GOPATH}/src \
   -I${GOPATH}/src/github.com/googleapis/googleapis \
   -I${GOPATH}/src/github.com/gogo/protobuf \
   --go_out=plugins=grpc:. \
   --go_opt=paths=source_relative \
   --grpc-gateway_out=. \
   --grpc-gateway_opt=paths=source_relative \
   --grpc-gateway_opt=logtostderr=true \
   --golug_out=. \
	example/proto/login/*.proto

.PHONY: clear
clear:
	rm -rf example/proto/*.go
	rm -rf example/proto/**/*.go

.PHONY: gen
gen:
	cd cmds/protoc-gen-golug && go install .

.PHONY: example
example:
	go build ${LDFLAGS} -mod vendor -v -o main example/*.go

.PHONY: run
run:
	go run ${LDFLAGS} -mod vendor -v example/*.go http

docker:
	docker build -t golug .

build-all:
	go build -tags "kcp quic" ./...

deps:
	go list -f '{{ join .Deps  "\n"}}' ./... |grep "/" | grep -v "github.com/smallnest/rpcx"| grep "\." | sort |uniq

vet:
	go vet ./...

tools:
	go install \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/golang/lint/golint \

cover:
	gocov test -tags "kcp quic" ./... | gocov-html > cover.html
	open cover.html

check-libs:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go list -m -u all | column -t

update-libs:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go get -u -v ./...

mod-tidy:
	GIT_TERMINAL_PROMPT=1 GO111MODULE=on go mod tidy

tools:
	@echo "libprotoc 3.11.4"
	go install \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/bufbuild/buf/cmd/buf \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
        github.com/golang/protobuf/protoc-gen-go \
        golang.org/x/tools/cmd/stringer \

mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o goctl-darwin goctl.go
	$(if $(shell command -v upx), upx goctl-darwin)
win:
	GOOS=windows go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o goctl.exe goctl.go
	$(if $(shell command -v upx), upx goctl.exe)
linux:
	GOOS=linux go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o goctl-linux goctl.go
	$(if $(shell command -v upx), upx goctl-linux)

changelog:
	docker run --rm \
		--interactive \
		--tty \
		-e "CHANGELOG_GITHUB_TOKEN=${CHANGELOG_GITHUB_TOKEN}" \
		-v "$(PWD):/usr/local/src/your-app" \
		ferrarimarco/github-changelog-generator:1.14.3 \
				-u grpc-ecosystem \
				-p grpc-gateway \
				--author \
				--compare-link \
				--github-site=https://github.com \
				--unreleased-label "**Next release**" \
				--release-branch=master \
				--future-release=v2.3.0