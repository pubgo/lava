package encoding

var Name = "encoding"

// Codec defines the interface.
type Codec interface {
	Name() string
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}
