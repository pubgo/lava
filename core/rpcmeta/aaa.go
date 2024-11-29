package rpcmeta

import "google.golang.org/protobuf/proto"

type RpcMeta struct {
	Method string
	Name   string
	Tags   map[string]string
	Input  proto.Message
	Output proto.Message
}
