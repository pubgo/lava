package watcher

import (
	"context"

	"github.com/pubgo/x/typex"
)

var _ Watcher = (*nullWatcher)(nil)

func init() {
	Register("noop", func(cfg typex.M) (Watcher, error) { return new(nullWatcher), nil })
}

type nullWatcher struct{}

func (e *nullWatcher) Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response { return nil }
func (e *nullWatcher) Name() string                                                        { return "noop" }
