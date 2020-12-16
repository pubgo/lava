// Code generated by protoc-gen-golug. DO NOT EDIT.
// source: example/proto/hello/transport.proto

package hello

import (
	"reflect"

	"github.com/pubgo/golug/golug_xgen"
)

func init() {
	var mthList []golug_xgen.GrpcRestHandler
	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Name:          "TestStream",
		Method:        "POST",
		Path:          "/hello_transport/test_stream",
		ClientStream:  "True" == "True",
		ServerStreams: "True" == "True",
	})

	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Name:          "TestStream1",
		Method:        "POST",
		Path:          "/hello_transport/test_stream1",
		ClientStream:  "True" == "True",
		ServerStreams: "False" == "True",
	})

	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Name:          "TestStream2",
		Method:        "POST",
		Path:          "/hello_transport/test_stream2",
		ClientStream:  "False" == "True",
		ServerStreams: "True" == "True",
	})

	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Name:          "TestStream3",
		Method:        "POST",
		Path:          "/hello_transport/test_stream3",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	golug_xgen.Add(reflect.ValueOf(RegisterTransportServer), mthList)
}
