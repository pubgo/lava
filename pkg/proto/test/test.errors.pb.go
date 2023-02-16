// Code generated by protoc-gen-lava-errors. DO NOT EDIT.
// versions:
// - protoc-gen-lava-errors v0.0.2
// - protoc                 v3.19.4
// source: proto/test/test.proto

package testpbv1

import (
	errors "github.com/pubgo/funk/errors"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

var ErrTestOK = errors.WrapBizCode(errors.WrapReason(errors.New("ok"), "ok"), "lava.test.v1.test.ok")
var ErrTestNotFound = errors.WrapBizCode(errors.WrapReason(errors.New("NotFound 找不到"), "NotFound 找不到"), "lava.test.v1.test.not_found")
var ErrTestUnknown = errors.WrapBizCode(errors.WrapReason(errors.New("Unknown 未知"), "Unknown 未知"), "lava.test.v1.test.unknown")
var ErrTestDbConn = errors.WrapBizCode(errors.WrapReason(errors.New("db connect error"), "db connect error"), "lava.test.v1.test.db_conn")