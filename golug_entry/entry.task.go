package golug_entry

import "context"

type Message struct {
	ID        []byte
	Body      []byte
	Timestamp int64
	Attempts  uint16
	Header    map[string]string
}

type TaskOptions struct {
	Ctx context.Context
}

type TaskOption func(opts *TaskOptions)
type TaskHandler func(topic string, data *Message) error
type TaskEntry interface {
	Entry
	Register(handler TaskHandler, opts ...TaskOption) error
}
