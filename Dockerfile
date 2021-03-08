FROM golang:1.16.0-alpine3.12 as builder

LABEL maintainer="barry"

ENV GO111MODULE  on
ENV GOPROXY      http://goproxy.cn
ENV GOSUMDB      sum.golang.google.cn
ENV GOPATH       /data/src

WORKDIR /data/src
COPY . /data/src/

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && apk update && apk upgrade \
    && apk add --no-cache bash protobuf curl git python3 openssh make gcc libc-dev \
    && rm -rf /var/cache/apk/* /tmp/*

# protoc *.h
RUN export PROTOC_ZIP=protoc-3.7.1-linux-x86_64.zip \
    && curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/$PROTOC_ZIP \
    && unzip -o $PROTOC_ZIP -d /usr/local 'include/*' \
    && rm -f $PROTOC_ZIP

# lint
ARG GOLANGCI_VERSION="v1.27.0"
RUN curl https://bootstrap.pypa.io/get-pip.py | python3 \
    && pip install yamllint==1.23.0 \
    && curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b "$GOPATH"/bin "$GOLANGCI_VERSION"

RUN go install -v github.com/golang/protobuf/protoc-gen-go \
    && go install -v github.com/gogo/protobuf/protoc-gen-gofast \
    && go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
    && go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    && go install -v github.com/vektra/mockery/cmd/mockery \
    && make install \
    && go clean -i -cache

FROM alpine:3.12.0

COPY --from=builder /go/bin/* /usr/local/bin/*
