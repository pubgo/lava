// Code generated by protoc-gen-golug. DO NOT EDIT.
// source: example/proto/hello/api.proto

package hello

import (
	"reflect"

	"github.com/pubgo/golug/service/grpclient"
	"github.com/pubgo/golug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func GetTestApiClient(srv string, opts ...grpc.DialOption) func() (TestApiClient, error) {
	client := grpclient.Client(srv, opts...)
	return func() (TestApiClient, error) {
		c, err := client.Get()
		return &testApiClient{c}, xerror.WrapF(err, "srv: %s", srv)
	}
}

func GetTestApiV2Client(srv string, opts ...grpc.DialOption) func() (TestApiV2Client, error) {
	client := grpclient.Client(srv, opts...)
	return func() (TestApiV2Client, error) {
		c, err := client.Get()
		return &testApiV2Client{c}, xerror.WrapF(err, "srv: %s", srv)
	}
}

func init() {

	var mthList []xgen.GrpcRestHandler

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:       "hello.TestApi",
		Name:          "Version",
		Method:        "POST",
		Path:          "/hello/test_api/version",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:       "hello.TestApi",
		Name:          "VersionTest",
		Method:        "GET",
		Path:          "/v1/example/versiontest",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	xgen.Add(reflect.ValueOf(RegisterTestApiServer), mthList)

	xgen.Add(reflect.ValueOf(RegisterTestApiHandlerFromEndpoint), nil)

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:       "hello.TestApiV2",
		Name:          "Version1",
		Method:        "POST",
		Path:          "/v2/example/version",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:       "hello.TestApiV2",
		Name:          "VersionTest1",
		Method:        "POST",
		Path:          "/v2/example/versiontest",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	xgen.Add(reflect.ValueOf(RegisterTestApiV2Server), mthList)

	xgen.Add(reflect.ValueOf(RegisterTestApiV2HandlerFromEndpoint), nil)

}
