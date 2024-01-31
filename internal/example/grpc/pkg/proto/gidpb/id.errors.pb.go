// Code generated by protoc-gen-go-errors. DO NOT EDIT.
// versions:
// - protoc-gen-go-errors v0.0.5
// - protoc                 v4.25.2
// source: gid/id.proto

package gidpb

import (
	errors "github.com/pubgo/funk/errors"
	errorpb "github.com/pubgo/funk/proto/errorpb"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

var ErrCodeOK = &errorpb.ErrCode{
	Code:       int32(0),
	Message:    "ok",
	Name:       "gid.ok",
	StatusCode: errorpb.Code_OK,
}
var _ = errors.RegisterErrCodes(ErrCodeOK)

var ErrCodeIDGenerateFailed = &errorpb.ErrCode{
	Code:       int32(100),
	Message:    "id generate error",
	Name:       "gid.id_generate_failed",
	StatusCode: errorpb.Code_Internal,
}
var _ = errors.RegisterErrCodes(ErrCodeIDGenerateFailed)
