package encoding

var Name = "encoding"

// Codec 定义了一个编解码器需要实现的接口。
// Encode/Decode 与 Marshal/Unmarshal 在多数实现中语义一致，
// 同时保留两组方法以兼容不同调用约定。
type Codec interface {
	// Name 返回编解码器名称，如 "json"、"proto"。
	Name() string
	// Encode 将对象序列化为字节切片。
	Encode(v any) ([]byte, error)
	// Decode 将字节切片反序列化到 v 指向的对象。
	Decode(data []byte, v any) error
	// Marshal 将对象序列化为字节切片。
	Marshal(v any) ([]byte, error)
	// Unmarshal 将字节切片反序列化到 v 指向的对象。
	Unmarshal(data []byte, v any) error
}
