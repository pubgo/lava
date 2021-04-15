package etcdv3

import (
	"context"
	"sync"

	"github.com/pubgo/lug/client/etcdv3"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/x/typex"
	"go.etcd.io/etcd/clientv3"
)

var Name = "etcd"

func init() {
	watcher.Register(Name, func(cfg typex.M) (watcher.Watcher, error) {
		return newWatcher("", ""), nil
	})
}

var _ watcher.Watcher = (*etcdWatcher)(nil)

func newWatcher(prefix string, name string) watcher.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &etcdWatcher{
		name:   name,
		prefix: prefix,
		ctx:    ctx,
		cancel: cancel,
		exitCh: make(chan struct{}, 1),
	}
}

type etcdWatcher struct {
	cancel context.CancelFunc

	name string

	mu     sync.Mutex
	client *clientv3.Client

	ctx context.Context

	closed   bool
	prefix   string
	revision int64
	exitCh   chan struct{}
}

func (w *etcdWatcher) Watch(ctx context.Context, key string, opts ...watcher.Opt) <-chan *watcher.Response {
	var resp = make(chan *watcher.Response)
	go func() {
		for w := range etcdv3.Get().Watch(ctx, key) {
			for i := range w.Events {
				var e = w.Events[i]
				resp <- &watcher.Response{
					Event:    e.Type.String(),
					Key:      string(e.Kv.Key),
					Value:    e.Kv.Value,
					Revision: e.Kv.Version,
				}
			}
		}
	}()

	return resp
}

func (w *etcdWatcher) Name() string {
	return w.prefix
}
