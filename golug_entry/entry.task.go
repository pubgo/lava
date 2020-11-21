package golug_entry

import (
	"context"
)

type Message struct {
	ID        []byte
	Body      []byte
	Timestamp int64
	Attempts  uint16
}

type TaskOptions struct{}
type TaskOption func(opts *TaskOptions)
type TaskHandler func(ctx context.Context, data *Message) error
type TaskEntry interface {
	Entry
	Register(name string, handler TaskHandler, opts ...TaskOption)
}
