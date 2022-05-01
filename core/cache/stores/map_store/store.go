package map_store

import (
	"github.com/pubgo/lava/core/cache"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"time"
)

const Name = "map"

func init() {
	cache.Register(Name, func(cfgMap map[string]interface{}) (cache.IStore, error) {
		var cfg Cfg

		if cfgMap != nil {
			xerror.Panic(merge.MapStruct(&cfg, cfgMap))
		}

		return &storeImpl{cfg: cfg}, nil
	})
}

type item struct {
	value   []byte
	expired time.Time
}

var _ cache.IStore = (*storeImpl)(nil)

type storeImpl struct {
	cfg      Cfg
	data     map[string]item
	callback func(key string, obj []byte)
}

func (s *storeImpl) GetExpired(key string) (obj []byte, expired time.Time, err error) {
	var dt, ok = s.data[key]
	if !ok {
		return nil, time.Time{}, cache.ErrNotFound
	}

	return dt.value, dt.expired, nil
}

func (s *storeImpl) OnEvicted(f func(key string, obj []byte)) {
	s.callback = f
}

func (s *storeImpl) Get(key string) (obj []byte, err error) {
	var dt, ok = s.data[key]
	if !ok {
		return nil, cache.ErrNotFound
	}

	return dt.value, nil
}

func (s *storeImpl) Set(key string, obj []byte, dur ...time.Duration) error {
	s.data[key] = item{value: obj, expired: time.Now().Add(dur[0])}
	return nil
}

func (s *storeImpl) Delete(key string) error {
	var dt, ok = s.data[key]
	if ok {
		delete(s.data, key)
		s.callback(key, dt.value)
	}

	return nil
}

func (s *storeImpl) DeleteExpired() error {
	for k, v := range s.data {
		if time.Since(v.expired) < 0 {
			continue
		}

		delete(s.data, k)
		s.callback(k, v.value)
	}

	return nil
}

func (s *storeImpl) Close() error { return nil }
