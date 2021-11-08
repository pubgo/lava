package etcdv3

import (
	"context"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/pubgo/lava/clients/etcdv3"
	"github.com/pubgo/lava/pkg/watcher"
	"github.com/pubgo/lava/types"
)

func init() {
	watcher.Register(Name, func(cfg types.M) (watcher.Watcher, error) {
		var c Cfg
		xerror.Exit(merge.MapStruct(&c, cfg))
		return newWatcher(c.Prefix, c.Name), nil
	})
}

var _ watcher.Watcher = (*watcherImpl)(nil)

func newWatcher(prefix string, name string) watcher.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcherImpl{
		name:   name,
		prefix: prefix,
		ctx:    ctx,
		cancel: cancel,
		exitCh: make(chan struct{}),
	}
}

type watcherImpl struct {
	ctx      context.Context
	name     string
	cancel   context.CancelFunc
	prefix   string
	revision int64
	exitCh   chan struct{}
}

func (w *watcherImpl) getEtcd() *etcdv3.Client {
	var cli = etcdv3.Get(w.name)
	xerror.Assert(cli == nil, "etcd client is nil")
	return cli
}

func (w *watcherImpl) Close(ctx context.Context, opts ...watcher.Opt) { close(w.exitCh) }
func (w *watcherImpl) Get(ctx context.Context, key string, opts ...watcher.Opt) (responses []*watcher.Response, gErr error) {
	defer xerror.RespErr(&gErr)

	key = handleKey(key)

	var resp, err = w.getEtcd().Get(ctx, key)
	xerror.Panic(err)

	for i := range resp.Kvs {
		e := resp.Kvs[i]
		responses = append(responses, &watcher.Response{
			Event:   types.EventType_UPDATE,
			Key:     string(e.Key),
			Value:   e.Value,
			Version: e.Version,
		})
	}
	return
}

func (w *watcherImpl) GetCallback(ctx context.Context, key string, fn func(resp *watcher.Response), opts ...watcher.Opt) (err error) {
	defer xerror.RespErr(&err)

	responses, err := w.Get(ctx, key, opts...)
	if err != nil {
		return xerror.Wrap(err)
	}

	for i := range responses {
		fn(responses[i])
	}

	return nil
}

func (w *watcherImpl) WatchCallback(ctx context.Context, key string, fn func(resp *watcher.Response), opts ...watcher.Opt) {
	key = handleKey(key)

	go func() {
		for w := range w.getEtcd().Watch(ctx, key) {
			for i := range w.Events {
				var e = w.Events[i]
				fn(&watcher.Response{
					Event:   convert(e.Type),
					Key:     string(e.Kv.Key),
					Value:   e.Kv.Value,
					Version: e.Kv.Version,
				})
			}
		}
	}()
}

func (w *watcherImpl) Watch(ctx context.Context, key string, opts ...watcher.Opt) <-chan *watcher.Response {
	key = handleKey(key)

	var resp = make(chan *watcher.Response)
	go func() {
		for w := range etcdv3.Get().Watch(ctx, key) {
			for i := range w.Events {
				var e = w.Events[i]
				resp <- &watcher.Response{
					Event:   convert(e.Type),
					Key:     string(e.Kv.Key),
					Value:   e.Kv.Value,
					Version: e.Kv.Version,
				}
			}
		}
	}()

	return resp
}

func (w *watcherImpl) Name() string {
	return w.prefix
}

func convert(ty mvccpb.Event_EventType) types.EventType {
	switch ty {
	case mvccpb.DELETE:
		return types.EventType_DELETE
	case mvccpb.PUT:
		return types.EventType_UPDATE
	default:
		return types.EventType_UNKNOWN
	}
}
