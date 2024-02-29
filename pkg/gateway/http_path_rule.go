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
	HttpPath string

	RawHttpPath string

	HttpMethod string

	// /{ServiceName}/{MethodName}
	GrpcMethodName string

	// variables on path
	// a.b.c=>a_b_c
	Vars map[string]string
	// Vars map[string][]protoreflect.FieldDescriptor

	// body=[""|"*"|"name"]
	reqBody []protoreflect.FieldDescriptor

	// resp_body=[""|"*"]
	rspBody []protoreflect.FieldDescriptor

	// body="*" or body="field.name" or body="" for no body
	HasReqBody bool

	// response_body="*" or response_body="field.name" or response_body="" for no body
	HasRspBody bool

	IsGroup bool
}
