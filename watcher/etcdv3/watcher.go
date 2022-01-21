package etcdv3

import (
	"context"

	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/pubgo/lava/clients/etcdv3"
	"github.com/pubgo/lava/event"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/watcher"
)

func init() {
	watcher.RegisterFactory(Name, func(cfg types.CfgMap) (watcher.Watcher, error) {
		var c Cfg
		merge.MapStruct(&c, cfg)
		return newWatcher(c.Prefix, c.Name), nil
	})
}

var _ watcher.Watcher = (*watcherImpl)(nil)

func newWatcher(prefix string, name string) watcher.Watcher {
	var cli = etcdv3.Get(name)
	xerror.Assert(cli == nil, "etcd client is nil")

	ctx, cancel := context.WithCancel(context.Background())
	return &watcherImpl{
		name:    name,
		prefix:  prefix,
		ctx:     ctx,
		cancel:  cancel,
		exitCh:  make(chan struct{}),
		etcdCli: cli,
	}
}

type watcherImpl struct {
	ctx      context.Context
	name     string
	cancel   context.CancelFunc
	prefix   string
	revision int64
	exitCh   chan struct{}
	etcdCli  *etcdv3.Client
}

func (w *watcherImpl) Close(ctx context.Context, opts ...watcher.Opt) { close(w.exitCh) }
func (w *watcherImpl) Get(ctx context.Context, key string, opts ...watcher.Opt) (responses []*watcher.Response, gErr error) {
	defer xerror.RespErr(&gErr)

	key = handleKey(key)

	var resp, err = w.etcdCli.Get().Get(ctx, key)
	xerror.Panic(err)

	for i := range resp.Kvs {
		e := resp.Kvs[i]
		responses = append(responses, &watcher.Response{
			Event:   event.EventType_UPDATE,
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
		for w := range w.etcdCli.Get().Watch(ctx, key) {
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
		for w := range w.etcdCli.Get().Watch(ctx, key) {
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

func convert(ty mvccpb.Event_EventType) event.EventType {
	switch ty {
	case mvccpb.DELETE:
		return event.EventType_DELETE
	case mvccpb.PUT:
		return event.EventType_UPDATE
	default:
		return event.EventType_UNKNOWN
	}
}
