package protojson

import (
	"bytes"
	"encoding/json"
	"github.com/pubgo/lava/encoding"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const Name = "jsonpb"

func init() {
	encoding.Register(Name, &jsonCodec{})
}

var useNumber bool

// UseNumber fix unmarshal Number(8234567890123456789) to interface(8.234567890123457e+18)
func UseNumber() {
	useNumber = true
}

var jsonpbMarshaler = &protojson.MarshalOptions{EmitUnpopulated: true}
var jsonpbUnmarshaler = &protojson.UnmarshalOptions{AllowPartial: true}

type jsonCodec struct{}

func (j *jsonCodec) Name() string { return Name }

func (j *jsonCodec) Encode(v interface{}) ([]byte, error) {
	if m, ok := v.(json.Marshaler); ok {
		return m.MarshalJSON()
	}

	if pb, ok := v.(proto.Message); ok {
		return jsonpbMarshaler.Marshal(pb)
	}

	return json.Marshal(v)
}

func (j *jsonCodec) Decode(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}

	if m, ok := v.(json.Unmarshaler); ok {
		return m.UnmarshalJSON(data)
	}

	if pb, ok := v.(proto.Message); ok {
		return jsonpbUnmarshaler.Unmarshal(data, pb)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if useNumber {
		dec.UseNumber()
	}
	
	return dec.Decode(v)
}

func (j *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return j.Encode(v)
}

func (j *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return j.Decode(data, v)
}
