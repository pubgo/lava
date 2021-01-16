// Code generated by protoc-gen-golug. DO NOT EDIT.
// source: example/proto/hello/api.proto

package hello

import (
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_client/grpclient"
	"github.com/pubgo/golug/golug_xgen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	var mthList []golug_xgen.GrpcRestHandler
	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Service:       "hello.TestApi",
		Name:          "Version",
		Method:        "POST",
		Path:          "/hello/test_api/version",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Service:       "hello.TestApi",
		Name:          "VersionTest",
		Method:        "GET",
		Path:          "/v1/example/versiontest",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	golug_xgen.Add(reflect.ValueOf(RegisterTestApiServer), mthList)
	golug_xgen.Add(reflect.ValueOf(RegisterTestApiGateway), struct{}{})
}

func init() {
	var mthList []golug_xgen.GrpcRestHandler
	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Service:       "hello.TestApiV2",
		Name:          "Version1",
		Method:        "POST",
		Path:          "/v2/example/version",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	mthList = append(mthList, golug_xgen.GrpcRestHandler{
		Service:       "hello.TestApiV2",
		Name:          "VersionTest1",
		Method:        "POST",
		Path:          "/v2/example/versiontest",
		ClientStream:  "False" == "True",
		ServerStreams: "False" == "True",
	})

	golug_xgen.Add(reflect.ValueOf(RegisterTestApiV2Server), mthList)
	golug_xgen.Add(reflect.ValueOf(RegisterTestApiV2Gateway), struct{}{})
}

func GetTestApiClient(srv string, opts ...grpc.DialOption) (TestApiClient, error) {
	c, err := grpclient.New(srv, opts...)
	return &testApiClient{c}, err
}

func GetTestApiV2Client(srv string, opts ...grpc.DialOption) (TestApiV2Client, error) {
	c, err := grpclient.New(srv, opts...)
	return &testApiV2Client{c}, err
}

func RegisterTestApiGateway(srv string, g fiber.Group, opts ...grpc.DialOption) error {
	c, err := GetTestApiClient(srv, opts...)
	if err != nil {
		return err
	}
	g.Add("POST", "/hello/test_api/version", func(ctx *fiber.Ctx) error {
		p := metadata.Pairs()
		ctx.Request().Header.VisitAll(func(key, value []byte) { p.Set(string(key), string(value)) })

		var req TestReq
		if err := ctx.BodyParser(&req); err != nil {
			return err
		}

		resp, err := c.Version(metadata.NewIncomingContext(ctx.Context(), p), req)
		return ctx.JSON(resp)
	})

	g.Add("GET", "/v1/example/versiontest", func(ctx *fiber.Ctx) error {
		p := metadata.Pairs()
		ctx.Request().Header.VisitAll(func(key, value []byte) { p.Set(string(key), string(value)) })

		var req TestReq
		var data = make(map[string]interface{})
		ctx.Context().QueryArgs().VisitAll(func(key, value []byte) { data[string(key)] = string(value) })
		if err := golug_utils.Decode(data, &req); err != nil {
			return err
		}

		resp, err := c.VersionTest(metadata.NewIncomingContext(ctx.Context(), p), req)
		return ctx.JSON(resp)
	})

}

func RegisterTestApiV2Gateway(srv string, g fiber.Group, opts ...grpc.DialOption) error {
	c, err := GetTestApiV2Client(srv, opts...)
	if err != nil {
		return err
	}
	g.Add("POST", "/v2/example/version", func(ctx *fiber.Ctx) error {
		p := metadata.Pairs()
		ctx.Request().Header.VisitAll(func(key, value []byte) { p.Set(string(key), string(value)) })

		var req TestReq
		if err := ctx.BodyParser(&req); err != nil {
			return err
		}

		resp, err := c.Version1(metadata.NewIncomingContext(ctx.Context(), p), req)
		return ctx.JSON(resp)
	})

	g.Add("POST", "/v2/example/versiontest", func(ctx *fiber.Ctx) error {
		p := metadata.Pairs()
		ctx.Request().Header.VisitAll(func(key, value []byte) { p.Set(string(key), string(value)) })

		var req TestReq
		if err := ctx.BodyParser(&req); err != nil {
			return err
		}

		resp, err := c.VersionTest1(metadata.NewIncomingContext(ctx.Context(), p), req)
		return ctx.JSON(resp)
	})

}
