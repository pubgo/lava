package cache

import (
	"time"
)

var _ IStore = (*noopStore)(nil)

type noopStore struct{}

func (n noopStore) OnEvicted(f func(key string, obj []byte)) {
	panic("implement me")
}

func (n noopStore) Get(key string) (obj []byte, err error) {
	panic("implement me")
}

func (n noopStore) GetExpired(key string) (obj []byte, expired time.Time, err error) {
	panic("implement me")
}

func (n noopStore) Set(key string, obj []byte, dur ...time.Duration) error {
	panic("implement me")
}

func (n noopStore) Delete(key string) error {
	panic("implement me")
}

func (n noopStore) DeleteExpired() error {
	panic("implement me")
}

func (n noopStore) Close() error {
	panic("implement me")
}
