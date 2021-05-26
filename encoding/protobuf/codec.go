package protobuf

import (
	proto "github.com/gogo/protobuf/proto"
	"github.com/pubgo/lug/encoding"
	"github.com/pubgo/xerror"
	pb "google.golang.org/protobuf/proto"
)

var Name = "protobuf"

func init() {
	encoding.Register(Name, protobufCodec{})
}

// protobufCodec uses protobuf marshaler and unmarshaler.
type protobufCodec struct{}

func (c protobufCodec) Marshal(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Marshaler); ok {
		return m.Marshal()
	}

	if m, ok := v.(pb.Message); ok {
		return pb.Marshal(m)
	}

	return nil, xerror.Fmt("%T is not a proto.Marshaler", v)
}

func (c protobufCodec) Unmarshal(data []byte, v interface{}) error {
	if m, ok := v.(proto.Unmarshaler); ok {
		return m.Unmarshal(data)
	}

	if m, ok := v.(pb.Message); ok {
		return pb.Unmarshal(data, m)
	}

	return xerror.Fmt("%T is not a proto.Unmarshaler", v)
}

func (c protobufCodec) Name() string {
	return Name
}

// Encode encodes an object into slice of bytes.
func (c protobufCodec) Encode(i interface{}) ([]byte, error) {
	if m, ok := i.(proto.Marshaler); ok {
		return m.Marshal()
	}

	if m, ok := i.(pb.Message); ok {
		return pb.Marshal(m)
	}

	return nil, xerror.Fmt("%T is not a proto.Marshaler", i)
}

// Decode decodes an object from slice of bytes.
func (c protobufCodec) Decode(data []byte, i interface{}) error {
	if m, ok := i.(proto.Unmarshaler); ok {
		return m.Unmarshal(data)
	}

	if m, ok := i.(pb.Message); ok {
		return pb.Unmarshal(data, m)
	}

	return xerror.Fmt("%T is not a proto.Unmarshaler", i)
}
