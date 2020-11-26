package golug_cron

import (
	"context"
	"sync"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/robfig/cron/v3"
)

type Event struct{ context.Context }
type CallBack func(event interface{}) error

var EmptyEntry = cron.Entry{}

type cronManager struct {
	cron *cron.Cron
	data sync.Map
}

func (t *cronManager) Add(name string, spec string, cmd CallBack) (grr error) {
	defer xerror.RespErr(&grr)

	if cmd == nil {
		return xerror.New("CallBack is nil")
	}

	id, err := t.cron.AddFunc(spec, func() {
		xprocess.Go(func(ctx context.Context) error {
			return xerror.Wrap(cmd(Event{Context: ctx}))
		})
	})
	xerror.Panic(err)
	actual, loaded := t.data.LoadOrStore(name, id)

	if !loaded {
		return nil
	}
	t.cron.Remove(actual.(cron.EntryID))

	return nil
}

func (t *cronManager) Get(name string) cron.Entry {
	val, ok := t.data.Load(name)
	if ok {
		return t.cron.Entry(val.(cron.EntryID))
	}
	return EmptyEntry
}

func (t *cronManager) List() map[string]cron.Entry {
	var data = make(map[string]cron.Entry)
	t.data.Range(func(key, value interface{}) bool {
		data[key.(string)] = t.cron.Entry(value.(cron.EntryID))
		return true
	})
	return data
}

func (t *cronManager) Remove(name string) error {
	id, ok := t.data.Load(name)
	if !ok {
		return nil
	}

	t.cron.Remove(id.(cron.EntryID))
	t.data.Delete(name)
	return nil
}

func (t *cronManager) Start() error {
	t.cron.Start()
	return nil
}

func (t *cronManager) Stop() error {
	t.cron.Stop()
	return nil
}

func New(opts ...cron.Option) *cronManager {
	return &cronManager{
		cron: cron.New(opts...),
	}
}
