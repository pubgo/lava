package golug_entry

import (
	"context"
)

type Message struct {
	ID        []byte
	Body      []byte
	Timestamp int64
	Attempts  uint16
	Priority  uint8
	ReplyTo   string
	Header    map[string]string
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler TaskHandler) error
}

type TaskCallOptions struct {
	Topic    string
	Queue    string
	Ctx      context.Context
	AutoAck  bool
	Consumer Consumer
}

type TaskCallOption func(*TaskCallOptions)
type TaskHandler func(data *Message) error
type TaskEntry interface {
	Entry
	Register(topic string, handler TaskHandler, opts ...TaskCallOption) error
}
