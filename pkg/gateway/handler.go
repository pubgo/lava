// Copyright 2021 Edward McFarlane. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gateway

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type handlerFunc func(*muxOptions, grpc.ServerStream) error

type handler struct {
	desc    protoreflect.MethodDescriptor
	handler handlerFunc

	// '/{Service}/{Method}'
	method string
}
