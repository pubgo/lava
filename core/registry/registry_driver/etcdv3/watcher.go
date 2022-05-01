package etcdv3

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"

	"github.com/pubgo/lava/clients/etcdv3"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/event"
)

type Watcher struct {
	revision int64
	stop     chan struct{}
	w        clientV3.WatchChan
	client   *etcdv3.Client
	timeout  time.Duration
}

func newWatcher(r *Registry, timeout time.Duration, opts ...registry.WatchOpt) (registry.Watcher, error) {
	var wo registry.WatchOpts
	for _, o := range opts {
		o(&wo)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan struct{})

	go func() {
		<-stop
		cancel()
	}()

	watchPath := prefix
	if len(wo.Service) > 0 {
		watchPath = servicePath(prefix, wo.Service) + "/"
	}

	resp, err := r.Client.Get().Get(ctx, watchPath, clientV3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return &Watcher{
		revision: resp.Header.Revision,
		stop:     stop,
		w:        r.Client.Get().Watch(ctx, watchPath, clientV3.WithPrefix(), clientV3.WithPrevKV(), clientV3.WithRev(resp.Header.Revision)),
		client:   r.Client,
		timeout:  timeout,
	}, nil
}

func (w *Watcher) Next() (*registry.Result, error) {
	for resp := range w.w {
		if resp.Err() != nil {
			return nil, resp.Err()
		}

		if resp.CompactRevision > w.revision {
			w.revision = resp.CompactRevision
		}
		if resp.Header.GetRevision() > w.revision {
			w.revision = resp.Header.GetRevision()
		}

		for _, ev := range resp.Events {
			service := decode(ev.Kv.Value)
			var action event.EventType

			switch ev.Type {
			case clientV3.EventTypePut:
				if ev.IsCreate() {
					action = event.EventType_CREATE
				} else if ev.IsModify() {
					action = event.EventType_UPDATE
				}
			case clientV3.EventTypeDelete:
				action = event.EventType_DELETE

				// get service from prevKv
				service = decode(ev.PrevKv.Value)
			}

			if service == nil {
				continue
			}
			return &registry.Result{
				Action:  action,
				Service: service,
			}, nil
		}
	}

	return nil, registry.ErrWatcherStopped
}

func (w *Watcher) Stop() error {
	select {
	case <-w.stop:
		return nil
	default:
		close(w.stop)
	}
	return nil
}
