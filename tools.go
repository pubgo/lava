//go:build tools
// +build tools

package lava

import (
	_ "github.com/favadi/protoc-go-inject-tag"
	_ "github.com/fullstorydev/grpcurl/cmd/grpcurl"
	_ "github.com/go-bindata/go-bindata/v3/go-bindata"
	_ "github.com/gogo/protobuf/protoc-gen-gofast"
	_ "github.com/gogo/protobuf/protoc-gen-gogo"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/gordonklaus/ineffassign"

	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"

	_ "github.com/mwitkow/go-proto-validators/protoc-gen-govalidators"
	_ "github.com/tinylib/msgp"

	_ "golang.org/x/tools/cmd/goyacc"
	_ "golang.org/x/tools/go/packages/gopackages"

	_ "github.com/google/gops"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
