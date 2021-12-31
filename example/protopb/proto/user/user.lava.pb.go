// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/user/user.proto

package gid

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc "github.com/pubgo/lava/clients/grpcc"
	xgen "github.com/pubgo/lava/xgen"
	grpc "google.golang.org/grpc"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetUserClient(srv string, opts ...func(cfg *grpcc.Cfg)) UserClient {
	return &userClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &GenerateRequest{},
		Output:       &GenerateResponse{},
		Service:      "gid.User",
		Name:         "Generate",
		Method:       "POST",
		Path:         "/v1/id/generate",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TypesRequest{},
		Output:       &TypesResponse{},
		Service:      "gid.User",
		Name:         "Types",
		Method:       "GET",
		Path:         "/v1/id/types",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterUserServer, mthList)
	var registerUserGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterUserHandlerClient(ctx, mux, NewUserClient(conn))
	}
	xgen.Add(registerUserGrpcClient, nil)
}
func GetABitOfEverythingServiceClient(srv string, opts ...func(cfg *grpcc.Cfg)) ABitOfEverythingServiceClient {
	return &aBitOfEverythingServiceClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "Create",
		Method:       "POST",
		Path:         "/v1/example/a_bit_of_everything/{float_value}/{double_value}/{int64_value}/separator/{uint64_value}/{int32_value}/{fixed64_value}/{fixed32_value}/{bool_value}/{string_value=strprefix/*}/{uint32_value}/{sfixed32_value}/{sfixed64_value}/{sint32_value}/{sint64_value}/{nonConventionalNameValue}/{enum_value}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CreateBody",
		Method:       "POST",
		Path:         "/v1/example/a_bit_of_everything",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &CreateBookRequest{},
		Output:       &Book{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CreateBook",
		Method:       "POST",
		Path:         "/v1/{parent=publishers/*}/books",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UpdateBookRequest{},
		Output:       &Book{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "UpdateBook",
		Method:       "PATCH",
		Path:         "/v1/{book.name=publishers/*/books/*}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "Update",
		Method:       "PUT",
		Path:         "/v1/example/a_bit_of_everything/{uuid}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UpdateV2Request{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "UpdateV2",
		Method:       "PUT",
		Path:         "/v2/example/a_bit_of_everything/{abe.uuid}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "GetQuery",
		Method:       "GET",
		Path:         "/v1/example/a_bit_of_everything/query/{uuid}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverythingRepeated{},
		Output:       &ABitOfEverythingRepeated{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "GetRepeatedQuery",
		Method:       "GET",
		Path:         "/v1/example/a_bit_of_everything_repeated/{path_repeated_float_value}/{path_repeated_double_value}/{path_repeated_int64_value}/{path_repeated_uint64_value}/{path_repeated_int32_value}/{path_repeated_fixed64_value}/{path_repeated_fixed32_value}/{path_repeated_bool_value}/{path_repeated_string_value}/{path_repeated_bytes_value}/{path_repeated_uint32_value}/{path_repeated_enum_value}/{path_repeated_sfixed32_value}/{path_repeated_sfixed64_value}/{path_repeated_sint32_value}/{path_repeated_sint64_value}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "DeepPathEcho",
		Method:       "POST",
		Path:         "/v1/example/deep_path/{single_nested.name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &durationpb.Duration{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "NoBindings",
		Method:       "GET",
		Path:         "/v2/example/NoBindings",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "Timeout",
		Method:       "GET",
		Path:         "/v2/example/timeout",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "ErrorWithDetails",
		Method:       "GET",
		Path:         "/v2/example/errorwithdetails",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &MessageWithBody{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "GetMessageWithBody",
		Method:       "POST",
		Path:         "/v2/example/withbody/{id}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &Body{},
		Output:       &emptypb.Empty{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "PostWithEmptyBody",
		Method:       "POST",
		Path:         "/v2/example/postwithemptybody/{name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CheckGetQueryParams",
		Method:       "GET",
		Path:         "/v1/example/a_bit_of_everything/params/get/{single_nested.name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CheckNestedEnumGetQueryParams",
		Method:       "GET",
		Path:         "/v1/example/a_bit_of_everything/params/get/nested_enum/{single_nested.ok}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ABitOfEverything{},
		Output:       &ABitOfEverything{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CheckPostQueryParams",
		Method:       "POST",
		Path:         "/v1/example/a_bit_of_everything/params/post/{string_value}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &wrapperspb.StringValue{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "OverwriteResponseContentType",
		Method:       "GET",
		Path:         "/v2/example/overwriteresponsecontenttype",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &CheckStatusResponse{},
		Service:      "gid.ABitOfEverythingService",
		Name:         "CheckStatus",
		Method:       "GET",
		Path:         "/v1/example/checkStatus",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterABitOfEverythingServiceServer, mthList)
	var registerABitOfEverythingServiceGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterABitOfEverythingServiceHandlerClient(ctx, mux, NewABitOfEverythingServiceClient(conn))
	}
	xgen.Add(registerABitOfEverythingServiceGrpcClient, nil)
}
func GetCamelCaseServiceNameClient(srv string, opts ...func(cfg *grpcc.Cfg)) CamelCaseServiceNameClient {
	return &camelCaseServiceNameClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &emptypb.Empty{},
		Service:      "gid.camelCaseServiceName",
		Name:         "Empty",
		Method:       "GET",
		Path:         "/v2/example/empty",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterCamelCaseServiceNameServer, mthList)
	var registerCamelCaseServiceNameGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterCamelCaseServiceNameHandlerClient(ctx, mux, NewCamelCaseServiceNameClient(conn))
	}
	xgen.Add(registerCamelCaseServiceNameGrpcClient, nil)
}
func GetAnotherServiceWithNoBindingsClient(srv string, opts ...func(cfg *grpcc.Cfg)) AnotherServiceWithNoBindingsClient {
	return &anotherServiceWithNoBindingsClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &emptypb.Empty{},
		Service:      "gid.AnotherServiceWithNoBindings",
		Name:         "NoBindings",
		Method:       "GET",
		Path:         "/v2/another/no-bindings",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterAnotherServiceWithNoBindingsServer, mthList)
	var registerAnotherServiceWithNoBindingsGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterAnotherServiceWithNoBindingsHandlerClient(ctx, mux, NewAnotherServiceWithNoBindingsClient(conn))
	}
	xgen.Add(registerAnotherServiceWithNoBindingsGrpcClient, nil)
}
