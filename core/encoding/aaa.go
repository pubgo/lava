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
