package gateway

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	_ "google.golang.org/genproto/googleapis/api/httpbody"
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
