package protojson

import (
	"bytes"
	"encoding/json"

	"github.com/pubgo/lava/v2/core/encoding"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const Name = "jsonpb"

func init() {
	encoding.Register(Name, &Codec{})
}

var useNumber bool

// UseNumber fix unmarshal Number(8234567890123456789) to interface(8.234567890123457e+18)
func UseNumber() {
	useNumber = true
}

var (
	jsonMarshaller = &protojson.MarshalOptions{EmitUnpopulated: true}
	jsonUnmarshal  = &protojson.UnmarshalOptions{RecursionLimit: 20, DiscardUnknown: true}
)

var Default = &Codec{}

type Codec struct{}

func (j *Codec) Name() string { return Name }

func (j *Codec) Encode(v any) ([]byte, error) {
	if m, ok := v.(json.Marshaler); ok {
		return m.MarshalJSON()
	}

	if pb, ok := v.(proto.Message); ok {
		return jsonMarshaller.Marshal(pb)
	}

	return json.Marshal(v)
}

func (j *Codec) Decode(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}

	if m, ok := v.(json.Unmarshaler); ok {
		return m.UnmarshalJSON(data)
	}

	if pb, ok := v.(proto.Message); ok {
		return jsonUnmarshal.Unmarshal(data, pb)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if useNumber {
		dec.UseNumber()
	}

	return dec.Decode(v)
}

func (j *Codec) Marshal(v any) ([]byte, error) {
	return j.Encode(v)
}

func (j *Codec) Unmarshal(data []byte, v any) error {
	return j.Decode(data, v)
}
