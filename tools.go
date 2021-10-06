// +build tools

package lug

import (
	_ "github.com/favadi/protoc-go-inject-tag"
	_ "github.com/fullstorydev/grpcurl/cmd/grpcurl"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/mwitkow/go-proto-validators/protoc-gen-govalidators"
	_ "github.com/tinylib/msgp"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	// 无效分配检查
	_ "github.com/gordonklaus/ineffassign"
	_ "golang.org/x/tools/cmd/goyacc"
	_ "github.com/go-bindata/go-bindata/v3/go-bindata"
)
