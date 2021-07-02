// Code generated by protoc-gen-lug. DO NOT EDIT.
// source: example/proto/login/code.proto

package login

import (
	"reflect"
	"strings"

	"github.com/pubgo/lug/plugins/grpcc"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

var _ = strings.Trim

func GetCodeClient(srv string, optFns ...func(service string) []grpc.DialOption) func() (CodeClient, error) {
	client := grpcc.GetClient(srv, optFns...)
	return func() (CodeClient, error) {
		c, err := client.Get()
		return &codeClient{c}, xerror.WrapF(err, "srv: %s", srv)
	}
}

func init() {
	var mthList []xgen.GrpcRestHandler

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Code",
		Name:         "SendCode",
		Method:       "POST",
		Path:         "/user/code/send-code",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Code",
		Name:         "Verify",
		Method:       "POST",
		Path:         "/user/code/verify",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Code",
		Name:         "IsCheckImageCode",
		Method:       "POST",
		Path:         "/user/code/is-check-image-code",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Code",
		Name:         "VerifyImageCode",
		Method:       "POST",
		Path:         "/user/code/verify-image-code",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Code",
		Name:         "GetSendStatus",
		Method:       "POST",
		Path:         "/user/code/get-send-status",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
	})

	xgen.Add(reflect.ValueOf(RegisterCodeServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterCodeHandlerFromEndpoint), nil)
}