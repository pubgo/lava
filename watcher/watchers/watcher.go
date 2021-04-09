package watchers

import (
	"context"

	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/x/typex"
)

const Name = "null"

var _ watcher.Watcher = (*nullWatcher)(nil)

func init() {
	watcher.Register(Name, func(cfg typex.M) (watcher.Watcher, error) { return new(nullWatcher), nil })
}

type nullWatcher struct{}

func (e *nullWatcher) Watch(ctx context.Context, key string, opts ...watcher.Opt) <-chan *watcher.Response {
	return nil
}

func (e *nullWatcher) Name() string { return Name }
