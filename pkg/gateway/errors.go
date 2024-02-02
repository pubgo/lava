package gateway

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errNotFound = status.Error(codes.NotFound, "not found")
	errMethod   = status.Error(codes.InvalidArgument, "httpPathRule not allowed")
)
