package watcher

import (
	"context"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/watcher/watcher_type"
)

var _ watcher_type.Watcher = (*nullWatcher)(nil)

func init() {
	RegisterFactory("noop", func(cfg config_type.CfgMap) (watcher_type.Watcher, error) { return new(nullWatcher), nil })
}

type nullWatcher struct{}

func (e *nullWatcher) Init() {}

func (e *nullWatcher) Name() string { return "noop" }
func (e *nullWatcher) Get(ctx context.Context, key string, opts ...watcher_type.Opt) ([]*watcher_type.Response, error) {
	return nil, nil
}
func (e *nullWatcher) GetCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) error {
	return nil
}
func (e *nullWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) {
}
func (e *nullWatcher) Close() {}
func (e *nullWatcher) Watch(ctx context.Context, key string, opts ...watcher_type.Opt) <-chan *watcher_type.Response {
	return nil
}
