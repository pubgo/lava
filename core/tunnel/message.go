package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// MaxMessageSize 单条控制消息的最大字节数（长度前缀 + JSON 负载）。
// 用于防止对端发送超大长度前缀导致的内存耗尽（OOM）攻击。
const MaxMessageSize = 4 << 20 // 4 MiB

// messageHeaderSize 长度前缀字节数（uint32 大端）
const messageHeaderSize = 4

// WriteMessage 以「4 字节大端长度前缀 + JSON 负载」的格式写出一条消息。
func WriteMessage(w io.Writer, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if len(data) > MaxMessageSize {
		return fmt.Errorf("%w: message size %d exceeds limit %d", ErrInvalidMessage, len(data), MaxMessageSize)
	}

	header := make([]byte, messageHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(data)))

	if _, err := w.Write(header); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

// ReadMessage 读取一条「4 字节大端长度前缀 + JSON 负载」的消息。
// 当声明的长度超过 MaxMessageSize 时返回 ErrInvalidMessage，避免分配超大缓冲区。
func ReadMessage(r io.Reader) (*Message, error) {
	header := make([]byte, messageHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header)
	if length > MaxMessageSize {
		return nil, fmt.Errorf("%w: declared message size %d exceeds limit %d", ErrInvalidMessage, length, MaxMessageSize)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
