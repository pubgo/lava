package grpcutil

import "time"

const (
	DefaultContentType = "application/grpc"
	DefaultMaxMsgSize  = 1024 * 1024 * 4
	DefaultTimeout     = time.Second * 2
)
