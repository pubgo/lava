// +build tools

package tools

import (
	//_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	//_ "google.golang.org/protobuf/cmd/protoc-gen-go"

	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/rakyll/statik"
	_ "github.com/gogo/protobuf/proto"
)


//import (
//	_ "github.com/alexkohler/nakedret"
//	_ "github.com/chzchzchz/goword"
//	_ "github.com/coreos/license-bill-of-materials"
//	_ "github.com/gordonklaus/ineffassign"
//	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
//	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
//	_ "github.com/gyuho/gocovmerge"
//	_ "github.com/hexfusion/schwag"
//	_ "github.com/mdempsky/unconvert"
//	_ "github.com/mgechev/revive"
//	_ "go.etcd.io/protodoc"
//	_ "honnef.co/go/tools/cmd/staticcheck"
//	_ "mvdan.cc/unparam"
//	_ "github.com/mikefarah/yq/v3"
//)
