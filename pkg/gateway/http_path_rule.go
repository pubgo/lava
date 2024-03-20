package gateway

import (
	"fmt"
	"github.com/pubgo/funk/errors"
	"google.golang.org/grpc"
	"net/http"
	"strings"

	"github.com/pubgo/funk/assert"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type serviceWrap struct {
	serviceDesc *grpc.ServiceDesc
	ss          interface{}
}

type httpPathRule struct {
	srv        *serviceWrap
	opts       *muxOptions
	methodDesc *grpc.MethodDesc
	streamDesc *grpc.StreamDesc
	desc       protoreflect.MethodDescriptor

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

func (h httpPathRule) Handle(stream grpc.ServerStream) error {
	if h.methodDesc != nil {
		ctx := stream.Context()

		reply, err := h.methodDesc.Handler(ss, ctx, stream.RecvMsg, h.opts.unaryInterceptor)
		if err != nil {
			return errors.WrapCaller(err)
		}

		return errors.WrapCaller(stream.SendMsg(reply))
	} else {
		info := &grpc.StreamServerInfo{
			FullMethod:     string(h.desc.FullName()),
			IsClientStream: h.streamDesc.ClientStreams,
			IsServerStream: h.streamDesc.ServerStreams,
		}

		if h.opts.streamInterceptor != nil {
			return h.opts.streamInterceptor(ss, stream, info, h.streamDesc.Handler)
		} else {
			return h.streamDesc.Handler(ss, stream)
		}
	}
}

func getMethod(opts *muxOptions, rule *annotations.HttpRule, desc protoreflect.MethodDescriptor, name string) []*httpPathRule {
	if rule == nil {
		return nil
	}

	var pathUrl, method string
	switch v := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		method = http.MethodGet
		pathUrl = v.Get
	case *annotations.HttpRule_Put:
		method = http.MethodPut
		pathUrl = v.Put
	case *annotations.HttpRule_Post:
		method = http.MethodPost
		pathUrl = v.Post
	case *annotations.HttpRule_Delete:
		method = http.MethodDelete
		pathUrl = v.Delete
	case *annotations.HttpRule_Patch:
		method = http.MethodPatch
		pathUrl = v.Patch
	//case *annotations.HttpRule_Custom:
	//	method = strings.ToUpper(v.Custom.Kind)
	//	pathUrl = v.Custom.Path
	default:
		panic(fmt.Errorf("unsupported http rule pattern %v", v))
	}

	//if strings.Contains(pathUrl, ":") {
	//	log.Error().Any("path", pathUrl).Msg("grpc http rule pattern url should not contain ':'")
	//	return nil
	//}

	//assert.If(strings.Contains(pathUrl, ":"), "grpc http rule pattern url should not contain ':', path=%s", pathUrl)

	normalPath := strings.ReplaceAll(pathUrl, ".", "_")
	normalPath = strings.ReplaceAll(normalPath, "}", "")
	normalPath = strings.ReplaceAll(normalPath, "{", ":")

	compiler := assert.Must1(httprule.Parse(pathUrl))
	tp := compiler.Compile()
	pattern := assert.Must1(httprule.NewPattern(tp.Version, tp.OpCodes, tp.Pool, tp.Verb))

	// Method already registered.
	m := &httpPathRule{
		opts:           opts,
		desc:           desc,
		GrpcMethodName: name,
		RawHttpPath:    pathUrl,
		HttpMethod:     method,

		// TODO 未来需要调整
		Vars:     getPathVarMap(pathUrl),
		HttpPath: normalPath,
		IsGroup:  isGroup(pathUrl),
		Pattern:  &pattern,
	}

	switch rule.Body {
	case "*":
		m.HasReqBody = true
	case "":
		m.HasReqBody = false
	default:
		m.HasReqBody = true
		inputFieldDescriptors := desc.Input().Fields()
		m.reqBody = fieldPath(inputFieldDescriptors, strings.Split(rule.Body, ".")...)
	}

	if method == http.MethodGet {
		m.HasReqBody = false
	}

	m.HasRspBody = true
	if rule.ResponseBody != "" {
		outputFieldDescriptors := desc.Output().Fields()
		m.rspBody = fieldPath(outputFieldDescriptors, strings.Split(rule.ResponseBody, ".")...)
	}

	var data = []*httpPathRule{m}
	for _, addRule := range rule.AdditionalBindings {
		assert.If(len(addRule.AdditionalBindings) != 0, "nested rules are not allowed")
		for _, m1 := range getMethod(opts, addRule, desc, name) {
			if m1 == nil {
				continue
			}

			data = append(data, m1)
		}
	}
	return data
}
