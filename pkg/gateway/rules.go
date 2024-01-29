// Copyright 2021 Edward McFarlane. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	_ "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	_ "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

type param struct {
	val protoreflect.Value
	fds []protoreflect.FieldDescriptor
}

func parseParam(fds []protoreflect.FieldDescriptor, raw []byte) (param, error) {
	if len(fds) == 0 {
		return param{}, fmt.Errorf("zero field")
	}
	fd := fds[len(fds)-1]

	switch kind := fd.Kind(); kind {
	case protoreflect.BoolKind:
		var b bool
		if err := json.Unmarshal(raw, &b); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfBool(b)}, nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		var x int32
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfInt32(x)}, nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		var x int64
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfInt64(x)}, nil

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		var x uint32
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfUint32(x)}, nil

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		var x uint64
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfUint64(x)}, nil

	case protoreflect.FloatKind:
		var x float32
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfFloat32(x)}, nil

	case protoreflect.DoubleKind:
		var x float64
		if err := json.Unmarshal(raw, &x); err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfFloat64(x)}, nil

	case protoreflect.StringKind:
		return param{fds: fds, val: protoreflect.ValueOfString(string(raw))}, nil

	case protoreflect.BytesKind:
		enc := base64.StdEncoding
		if bytes.ContainsAny(raw, "-_") {
			enc = base64.URLEncoding
		}
		if len(raw)%4 != 0 {
			enc = enc.WithPadding(base64.NoPadding)
		}

		dst := make([]byte, enc.DecodedLen(len(raw)))
		n, err := enc.Decode(dst, raw)
		if err != nil {
			return param{}, err
		}
		return param{fds: fds, val: protoreflect.ValueOfBytes(dst[:n])}, nil

	case protoreflect.EnumKind:
		var x int32
		if err := json.Unmarshal(raw, &x); err == nil {
			return param{fds: fds, val: protoreflect.ValueOfEnum(protoreflect.EnumNumber(x))}, nil
		}

		s := string(raw)
		if isNullValue(fd) && s == "null" {
			return param{fds: fds, val: protoreflect.ValueOfEnum(0)}, nil
		}

		enumVal := fd.Enum().Values().ByName(protoreflect.Name(s))
		if enumVal == nil {
			return param{}, fmt.Errorf("unexpected enum %s", raw)
		}
		return param{fds: fds, val: protoreflect.ValueOfEnum(enumVal.Number())}, nil

	case protoreflect.MessageKind:
		// Well known JSON scalars are decoded to message types.
		md := fd.Message()
		name := string(md.FullName())
		if strings.HasPrefix(name, "google.protobuf.") {
			switch md.FullName()[16:] {
			case "Timestamp":
				var msg timestamppb.Timestamp
				if err := protojson.Unmarshal(quote(raw), &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "Duration":
				var msg durationpb.Duration
				if err := protojson.Unmarshal(quote(raw), &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "BoolValue":
				var msg wrapperspb.BoolValue
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "Int32Value":
				var msg wrapperspb.Int32Value
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "Int64Value":
				var msg wrapperspb.Int64Value
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "UInt32Value":
				var msg wrapperspb.UInt32Value
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "UInt64Value":
				var msg wrapperspb.UInt64Value
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "FloatValue":
				var msg wrapperspb.FloatValue
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "DoubleValue":
				var msg wrapperspb.DoubleValue
				if err := protojson.Unmarshal(raw, &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "BytesValue":
				var msg wrapperspb.BytesValue
				if err := protojson.Unmarshal(quote(raw), &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "StringValue":
				var msg wrapperspb.StringValue
				if err := protojson.Unmarshal(quote(raw), &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			case "FieldMask":
				var msg fieldmaskpb.FieldMask
				if err := protojson.Unmarshal(quote(raw), &msg); err != nil {
					return param{}, err
				}
				return param{fds: fds, val: protoreflect.ValueOfMessage(msg.ProtoReflect())}, nil
			}
		}
		return param{}, fmt.Errorf("unexpected message type %s", name)

	default:
		return param{}, fmt.Errorf("unknown param type %s", kind)

	}
}

func isNullValue(fd protoreflect.FieldDescriptor) bool {
	ed := fd.Enum()
	return ed != nil && ed.FullName() == "google.protobuf.NullValue"
}

type params []param

func (ps params) set(m proto.Message) error {
	for _, p := range ps {
		cur := m.ProtoReflect()
		for i, fd := range p.fds {
			if len(p.fds)-1 == i {
				switch {
				case fd.IsList():
					l := cur.Mutable(fd).List()
					l.Append(p.val)
				case fd.IsMap():
					return fmt.Errorf("map fields are not supported")
				default:
					cur.Set(fd, p.val)
				}
				break
			}

			cur = cur.Mutable(fd).Message()
		}
	}
	return nil
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
