package golug_watcher_etcd

import (
	"context"
	"strings"
	"sync"

	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"go.etcd.io/etcd/clientv3"
)

var _ golug_watcher.Watcher = (*etcdWatcher)(nil)

func newWatcher(prefix string, client *clientv3.Client) golug_watcher.Watcher {
	ctx, cancel := context.WithCancel(context.Background())

	resp, err := client.Get(context.Background(), prefix, clientv3.WithPrefix())
	xerror.Panic(err)

	return &etcdWatcher{
		revision: resp.Header.Revision,
		prefix:   prefix,
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		exitCh:   make(chan struct{}, 1),
	}
}

type etcdWatcher struct {
	cancel context.CancelFunc

	mu     sync.Mutex
	client *clientv3.Client

	ctx context.Context

	closed   bool
	prefix   string
	revision int64
	exitCh   chan struct{}
}

func (w *etcdWatcher) String() string {
	return "etcd"
}

func (w *etcdWatcher) Start() error {
	rch := w.client.Watch(context.Background(), w.prefix, clientv3.WithRev(w.revision+1), clientv3.WithPrefix())
	w.cancel = xprocess.GoLoop(func(ctx context.Context) error {
		resp, ok := <-rch
		if !ok {
			return xprocess.Break
		}

		if err := resp.Err(); err != nil {
			xlog.Error("etcdWatcher.Start handle error", xlog.Any("err", xerror.Parse(err)))
			return nil
		}

		var wg = xprocess.NewGroup()
		defer wg.Wait()
		for _, event := range resp.Events {
			key := handleKey(strings.TrimPrefix(handleKey(string(event.Kv.Key)), w.prefix))
			val := golug_watcher.GetCallBack(key)
			if val == nil {
				xlog.Errorf("[CallBack] is nil, name:%s", key)
				continue
			}

			wg.Go(func(ctx context.Context) {
				if err := val(&golug_watcher.Response{
					Event:    event.Type.String(),
					Key:      key,
					Value:    event.Kv.Value,
					Revision: event.Kv.ModRevision,
				}); err != nil {
					xlog.Errorf("%s handle error", xlog.Any("err", err))
				}
			})
		}
		return nil
	})
	return nil
}

// Close 关闭 etcdWatcher
func (w *etcdWatcher) Close() error {
	w.cancel()
	return nil
}

func handleKey(keys ...string) string {
	key := strings.Join(keys, ".")
	key = strings.ReplaceAll(key, "/", ".")
	key = strings.ReplaceAll(key, "..", ".")
	return strings.Trim(key, ".")
}
