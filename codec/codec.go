package codec

var Name = "codec"

// Codec defines the interface.
type Codec interface {
	Name() string
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}
