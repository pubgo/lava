package watcher

import (
	"context"

	"github.com/pubgo/lava/types"
)

var _ Watcher = (*nullWatcher)(nil)

func init() {
	Register("noop", func(cfg types.M) (Watcher, error) { return new(nullWatcher), nil })
}

type nullWatcher struct{}

func (e *nullWatcher) Get(ctx context.Context, key string, opts ...Opt) ([]*Response, error) {
	return nil, nil
}
func (e *nullWatcher) GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error {
	return nil
}
func (e *nullWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) {
}
func (e *nullWatcher) Close(ctx context.Context, opts ...Opt) {}
func (e *nullWatcher) Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response {
	return nil
}
func (e *nullWatcher) Name() string { return "noop" }
