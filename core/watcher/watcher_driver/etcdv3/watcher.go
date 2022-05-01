package etcdv3

import (
	"context"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/pubgo/lava/clients/etcdv3"
	"github.com/pubgo/lava/config"
	watcher2 "github.com/pubgo/lava/core/watcher"
	"github.com/pubgo/lava/event"
)

func init() {
	watcher2.RegisterFactory(Name, func(cfg config.CfgMap) (watcher2.Watcher, error) {
		var c Cfg
		xerror.Panic(cfg.Decode(&c))
		return newWatcher(c), nil
	})
}

var _ watcher2.Watcher = (*watcherImpl)(nil)

func newWatcher(cfg Cfg) watcher2.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcherImpl{
		Cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		exitCh: make(chan struct{}),
	}
}

type watcherImpl struct {
	Cfg        Cfg
	ctx        context.Context
	DriverName string
	cancel     context.CancelFunc
	revision   int64
	exitCh     chan struct{}
	Etcd       *etcdv3.Client `inject-expr:"Cfg.Name"`
}

func (w *watcherImpl) Init() {
	xerror.Assert(w.Etcd == nil, "etcd client is nil")
}

func (w *watcherImpl) Close() { close(w.exitCh) }
func (w *watcherImpl) Get(ctx context.Context, key string, opts ...watcher2.Opt) (responses []*watcher2.Response, gErr error) {
	defer xerror.RespErr(&gErr)

	key = handleKey(key)

	var resp, err = w.Etcd.Get().Get(ctx, key)
	xerror.Panic(err)

	for i := range resp.Kvs {
		e := resp.Kvs[i]
		responses = append(responses, &watcher2.Response{
			Event:   event.EventType_UPDATE,
			Key:     string(e.Key),
			Value:   e.Value,
			Version: e.Version,
		})
	}
	return
}

func (w *watcherImpl) GetCallback(ctx context.Context, key string, fn func(resp *watcher2.Response), opts ...watcher2.Opt) (err error) {
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

func (w *watcherImpl) WatchCallback(ctx context.Context, key string, fn func(resp *watcher2.Response), opts ...watcher2.Opt) {
	key = handleKey(key)

	go func() {
		for w := range w.Etcd.Get().Watch(ctx, key) {
			for i := range w.Events {
				var e = w.Events[i]
				fn(&watcher2.Response{
					Event:   convert(e.Type),
					Key:     string(e.Kv.Key),
					Value:   e.Kv.Value,
					Version: e.Kv.Version,
				})
			}
		}
	}()
}

func (w *watcherImpl) Watch(ctx context.Context, key string, opts ...watcher2.Opt) <-chan *watcher2.Response {
	key = handleKey(key)

	var resp = make(chan *watcher2.Response)
	go func() {
		for w := range w.Etcd.Get().Watch(ctx, key) {
			for i := range w.Events {
				var e = w.Events[i]
				resp <- &watcher2.Response{
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
