// Code generated by protoc-gen-lug. DO NOT EDIT.
// versions:
// - protoc-gen-lug v0.1.0
// - protoc         v3.17.3
// source: example/proto/hello/api.proto

package hello

import (
	grpcc "github.com/pubgo/lug/plugins/grpcc"
	xgen "github.com/pubgo/lug/xgen"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetTestApiClient(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli TestApiClient)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli TestApiClient)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&testApiClient{c})
		return
	}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApi",
		Name:         "Version",
		Method:       "GET",
		Path:         "/v1/version",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &structpb.Value{},
		Output:       &TestApiOutput1{},
		Service:      "hello.TestApi",
		Name:         "Version1",
		Method:       "POST",
		Path:         "/v1/version1",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApi",
		Name:         "VersionTest",
		Method:       "GET",
		Path:         "/v1/example/versiontest",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(reflect.ValueOf(RegisterTestApiServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterTestApiHandlerServer), nil)
}
func GetTestApiV2Client(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli TestApiV2Client)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli TestApiV2Client)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&testApiV2Client{c})
		return
	}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApiV2",
		Name:         "Version1",
		Method:       "POST",
		Path:         "/v2/example/version/{name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApiV2",
		Name:         "VersionTest1",
		Method:       "POST",
		Path:         "/v2/example/versiontest",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(reflect.ValueOf(RegisterTestApiV2Server), mthList)
	xgen.Add(reflect.ValueOf(RegisterTestApiV2HandlerServer), nil)
}