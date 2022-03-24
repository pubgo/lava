package noop

import (
	"context"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/core/watcher"
	"github.com/pubgo/lava/core/watcher/watcher_type"
)

var _ watcher_type.Watcher = (*NullWatcher)(nil)

func init() {
	watcher.RegisterFactory("noop", func(cfg config_type.CfgMap) (watcher_type.Watcher, error) { return new(NullWatcher), nil })
}

type NullWatcher struct{}

func (e *NullWatcher) Init() {}

func (e *NullWatcher) Name() string { return "noop" }
func (e *NullWatcher) Get(ctx context.Context, key string, opts ...watcher_type.Opt) ([]*watcher_type.Response, error) {
	return nil, nil
}
func (e *NullWatcher) GetCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) error {
	return nil
}
func (e *NullWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) {
}
func (e *NullWatcher) Close() {}
func (e *NullWatcher) Watch(ctx context.Context, key string, opts ...watcher_type.Opt) <-chan *watcher_type.Response {
	return nil
}
