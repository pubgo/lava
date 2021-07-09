package cache

import (
	"time"
)

var _ IStore = (*noopStore)(nil)

type noopStore struct{}

func (n noopStore) Get(key string) ([]byte, error) {
	panic("implement me")
}

func (n noopStore) GetObj(key string, obj interface{}) error {
	panic("implement me")
}

func (n noopStore) Set(key string, obj interface{}, ds ...time.Duration) error {
	panic("implement me")
}

func (n noopStore) GetSet(key string, d time.Duration, cbs ...getCallback) ([]byte, error) {
	panic("implement me")
}

func (n noopStore) GetExpired(key string) (obj []byte, expireAt int64, err error) {
	panic("implement me")
}

func (n noopStore) Delete(key string) error {
	panic("implement me")
}

func (n noopStore) DeleteExpired() error {
	panic("implement me")
}

func (n noopStore) OnEvicted(f func(key string, obj []byte)) {
	panic("implement me")
}

func (n noopStore) Close() error {
	panic("implement me")
}

