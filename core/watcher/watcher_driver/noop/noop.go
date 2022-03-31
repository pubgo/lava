package noop

import (
	"context"
	"github.com/pubgo/lava/config"

	"github.com/pubgo/lava/core/watcher"
)

var _ watcher.Watcher = (*NullWatcher)(nil)

func init() {
	watcher.RegisterFactory("noop", func(cfg config.CfgMap) (watcher.Watcher, error) { return new(NullWatcher), nil })
}

type NullWatcher struct{}

func (e *NullWatcher) Init() {}

func (e *NullWatcher) Name() string { return "noop" }
func (e *NullWatcher) Get(ctx context.Context, key string, opts ...watcher.Opt) ([]*watcher.Response, error) {
	return nil, nil
}
func (e *NullWatcher) GetCallback(ctx context.Context, key string, fn func(resp *watcher.Response), opts ...watcher.Opt) error {
	return nil
}
func (e *NullWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *watcher.Response), opts ...watcher.Opt) {
}
func (e *NullWatcher) Close() {}
func (e *NullWatcher) Watch(ctx context.Context, key string, opts ...watcher.Opt) <-chan *watcher.Response {
	return nil
}
