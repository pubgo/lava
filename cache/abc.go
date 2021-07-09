package cache

import (
	"time"
)

type getCallback func(key string) (interface{}, error)

type IStore interface {
	Get(key string) ([]byte, error)
	GetObj(key string, obj interface{}) error
	Set(key string, obj interface{}, ds ...time.Duration) error
	GetSet(key string, d time.Duration, cbs ...getCallback) ([]byte, error)
	GetExpired(key string) (obj []byte, expireAt int64, err error)
	Delete(key string) error
	DeleteExpired() error
	OnEvicted(func(key string, obj []byte))
	Close() error
}
