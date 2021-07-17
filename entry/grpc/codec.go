package grpc

import (
	"google.golang.org/grpc/encoding"
)

type jsonCodec struct{}
type bytesCodec struct{}
type protoCodec struct{}
type wrapCodec struct{ encoding.Codec }

var (
	defaultGRPCCodecs = map[string]encoding.Codec{
		//"application/json":         jsonCodec{},
		//"application/proto":        protoCodec{},
		//"application/protobuf":     protoCodec{},
		//"application/octet-stream": protoCodec{},
		//"application/grpc":         protoCodec{},
		//"application/grpc+json":    jsonCodec{},
		//"application/grpc+proto":   protoCodec{},
		//"application/grpc+bytes":   bytesCodec{},
	}
)

//func init() {
//	encoding.RegisterCodec(&uriCodec{})
//}
//
//// 解析http get请求的query参数
//type uriCodec struct{}
//
//func (c *uriCodec) Name() string { return "uri" }
//func (c *uriCodec) Marshal(v interface{}) ([]byte, error) {
//	return json.Marshal(v)
//}
//
//func (c *uriCodec) Unmarshal(data []byte, v interface{}) error {
//	var u, err = url.ParseQuery(string(data))
//	if err != nil {
//		return err
//	}
//
//	return gutil.MapFormByTag(v, u, "json")
//}
