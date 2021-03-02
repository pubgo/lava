FROM golang:1.14-alpine

ENV DRONE_REPO $DRONE_REPO
ENV GITLAB_PRIVATE_TOKEN $GITLAB_PRIVATE_TOKEN
ENV DRONE_REPO_BRANCH $DRONE_REPO_BRANCH

# 公司的go代理
ENV GOPROXY "http://goproxy.cn"
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
COPY micro-common-files /micro-common-files
RUN curl https://bootstrap.pypa.io/get-pip.py | python3 \
    && pip install yamllint==1.23.0 \
    && curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b "$GOPATH"/bin "$GOLANGCI_VERSION"

COPY ./vendor ./docker/proto ./docker/gosec ./docker/module ./docker/lint ./docker/swagger ./docker/changelog /go/src/
RUN go install -v github.com/golang/protobuf/protoc-gen-go \
    && go install -v github.com/gogo/protobuf/protoc-gen-gofast \
    && go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
    && go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    && go install -v github.com/micro/protoc-gen-micro \
    && go install -v github.com/vektra/mockery/cmd/mockery \
    && go install -v proto_option \
    && go install -v gather_module \
    && go install -v lint \
    && go install -v gosec \
    && go install -v changelog \
    && go install -v swagger_upload \
    && go install -v swagger_json \
    && go clean -i -cache
