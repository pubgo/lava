// Copyright 2021 Edward McFarlane. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gateway

import (
	"net/url"
	"strings"

	_ "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	_ "google.golang.org/protobuf/types/descriptorpb"
)

type httpPathRule struct {
	opts *muxOptions
	desc protoreflect.MethodDescriptor

	// Depth first search preferring path segments over variables.
	// Variables split the search tree:
	//
	//	/path/{variable/*}/to/{end/**} ?:VERB
	httpPath string

	rawHttpPath string

	httpMethod string

	// /{ServiceName}/{MethodName}
	grpcMethodName string

	// variables on path
	// a.b.c=>a_b_c
	vars map[string]string
	//vars map[string][]protoreflect.FieldDescriptor

	// body=[""|"*"|"name"]
	reqBody []protoreflect.FieldDescriptor

	// resp_body=[""|"*"]
	rspBody []protoreflect.FieldDescriptor

	// body="*" or body="field.name" or body="" for no body
	hasReqBody bool

	// response_body="*" or response_body="field.name" or response_body="" for no body
	hasRspBody bool

	isGroup bool
}

func (m *httpPathRule) String() string {
	return m.grpcMethodName
}

func (m *httpPathRule) parseQueryParams(values url.Values) (params, error) {
	msgDesc := m.desc.Input()
	fieldDescs := msgDesc.Fields()

	var ps params
	for key, vs := range values {
		fds := fieldPath(fieldDescs, strings.Split(key, ".")...)
		if fds == nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown query param %q", key)
		}

		for _, v := range vs {
			p, err := parseParam(fds, []byte(v))
			if err != nil {
				return nil, err
			}
			ps = append(ps, p)
		}
	}
	return ps, nil
}
