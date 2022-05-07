// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/hello/example.proto

package hellopb

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc_builder "github.com/pubgo/lava/clients/grpcc/grpcc_builder"
	inject "github.com/pubgo/lava/inject"
	service "github.com/pubgo/lava/service"
	xgen "github.com/pubgo/lava/xgen"
	fx "go.uber.org/fx"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func InitUserServiceClient(addr string, alias ...string) {

	var name = ""
	if len(alias) > 0 {
		name = alias[0]
	}
	conn := grpcc_builder.NewClient(addr)

	inject.Register(fx.Provide(fx.Annotated{
		Target: func() UserServiceClient { return NewUserServiceClient(conn) },
		Name:   name,
	}))
}

func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &User{},
		Output:       &emptypb.Empty{},
		Service:      "hello.UserService",
		Name:         "AddUser",
		Method:       "POST",
		Path:         "/hello/user-service/add-user",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &User{},
		Output:       &emptypb.Empty{},
		Service:      "hello.UserService",
		Name:         "GetUser",
		Method:       "POST",
		Path:         "/hello/user-service/get-user",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &ListUsersRequest{},
		Output:       &User{},
		Service:      "hello.UserService",
		Name:         "ListUsers",
		Method:       "POST",
		Path:         "/hello/user-service/list-users",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: true,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UserRole{},
		Output:       &User{},
		Service:      "hello.UserService",
		Name:         "ListUsersByRole",
		Method:       "POST",
		Path:         "/hello/user-service/list-users-by-role",
		DefaultUrl:   true,
		ClientStream: true,
		ServerStream: true,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UpdateUserRequest{},
		Output:       &User{},
		Service:      "hello.UserService",
		Name:         "UpdateUser",
		Method:       "POST",
		Path:         "/hello/user-service/update-user",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	xgen.Add(RegisterUserServiceServer, mthList)
}

func RegisterUserService(srv service.Service, impl UserServiceServer) {
	srv.RegService(service.Desc{
		Handler:     impl,
		ServiceDesc: UserService_ServiceDesc,
	})

	srv.RegGateway(func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error {
		return RegisterUserServiceHandlerClient(ctx, mux, NewUserServiceClient(cc))
	})

}
