// Code generated by protoc-gen-golug. DO NOT EDIT.
// source: example/proto/login/login.proto

package login

import (
	"reflect"

	"github.com/pubgo/golug/golug_client/grpclient"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_xgen"
)

func init() {
	var _mth []golug_entry.GrpcRestHandler

	_mth = append(_mth, golug_entry.GrpcRestHandler{
		Name:          "Login",
		Method:        "POST",
		Path:          "/user/login/login",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	_mth = append(_mth, golug_entry.GrpcRestHandler{
		Name:          "Authenticate",
		Method:        "POST",
		Path:          "/user/login/authenticate",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	golug_xgen.Add(reflect.ValueOf(RegisterLoginServer), _mth)
}

func GetLoginClient(srv string) LoginClient {
	return &loginClient{grpclient.GetClient(srv)}
}