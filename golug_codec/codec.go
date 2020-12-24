package golug_codec

// Codec defines the interface that decode/encode payload.
type Codec interface {
	Name() string
	Encode(i interface{}) ([]byte, error)
	Decode(data []byte, i interface{}) error
}
