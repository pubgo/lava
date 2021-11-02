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

	// Render markdown on the CLI, with pizzazz!
	_ "github.com/charmbracelet/glow"

	// An interactive web UI for gRPC, along the lines of postman
	_ "github.com/fullstorydev/grpcui/cmd/grpcui"

	// Simple gRPC benchmarking and load testing tool
	_ "github.com/bojand/ghz"

	// A simple zero-config tool to make locally trusted development certificates with any names you'd like.
	_ "filippo.io/mkcert"

	// Evans: more expressive universal gRPC client
	_ "github.com/ktr0731/evans"

	// Bit is a modern Git CLI
	_ "github.com/chriswalz/bit"

	// This is a simple tool to sign, verify and show JSON Web Tokens from the command line.
	_ "github.com/golang-jwt/jwt/v4/cmd/jwt"

	// This is a small reverse proxy that can front existing gRPC servers and expose their functionality using gRPC-Web protocol,
	//	allowing for the gRPC services to be consumed from browsers.
	_ "github.com/improbable-eng/grpc-web/go/grpcwebproxy"
)
