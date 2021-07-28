// +build tools

package lug

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/magefile/mage/mg"
	_ "github.com/magefile/mage/sh"
	_ "github.com/mwitkow/go-proto-validators/protoc-gen-govalidators"
	_ "github.com/pubgo/lug/tracing/jaeger"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
