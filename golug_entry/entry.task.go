package golug_entry

import (
	"context"
)

type TaskOptions struct{}
type TaskOption func(opts *TaskOptions)
type TaskHandler func(ctx context.Context, data []byte) error
type TaskEntry interface {
	Entry
	Register(name string, handler TaskHandler, opts ...TaskOption)
}
