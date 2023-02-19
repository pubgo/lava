package encoding

var Name = "encoding"

// Codec defines the interface.
type Codec interface {
	Name() string
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// GetWithCT get codec with content type
func GetWithCT(ct string) Codec {
	return Get(cdcMapping[ct])
}

var cdcMapping = map[string]string{
	"application/json":         "json",
	"application/proto":        "proto",
	"application/protobuf":     "proto",
	"application/octet-stream": "proto",
	"application/grpc":         "proto",
	"application/grpc+json":    "json",
	"application/grpc+proto":   "proto",
	"application/grpc+bytes":   "bytes",
}
