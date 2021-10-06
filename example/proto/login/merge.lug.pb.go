// Code generated by protoc-gen-lug. DO NOT EDIT.
// versions:
// - protoc-gen-lug v0.1.0
// - protoc         v3.17.3
// source: example/proto/login/merge.proto

package login

import (
	grpcc "github.com/pubgo/lug/plugins/grpcc"
	xgen "github.com/pubgo/lug/xgen"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetMergeClient(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli MergeClient)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli MergeClient)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&mergeClient{c})
		return
	}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "Telephone",
		Method:       "POST",
		Path:         "/user/merge/telephone",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "TelephoneCheck",
		Method:       "POST",
		Path:         "/user/merge/telephone-check",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChat",
		Method:       "POST",
		Path:         "/user/merge/we-chat",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatCheck",
		Method:       "POST",
		Path:         "/user/merge/we-chat-check",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatUnMergeRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatUnMerge",
		Method:       "POST",
		Path:         "/user/merge/we-chat-un-merge",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(reflect.ValueOf(RegisterMergeServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterMergeHandlerServer), nil)
}