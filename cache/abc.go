package cache

import (
	"time"
)

type getCallback func(key string) (interface{}, error)

type Store interface {
	Get(key string) ([]byte, error)
	GetObj(key string, o interface{}) error
	Set(key string, val interface{}, ds ...time.Duration) error
	GetSet(key string, d time.Duration, fns ...getCallback) (interface{}, error)
	GetExpired(key string) (val interface{}, expireAt int64, err error)
	Delete(key string) error
	DeleteExpired() error
	OnEvicted(func(key string, val interface{}))
	Close() error
}
