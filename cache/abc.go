package cache

import (
	"time"
)

type getCallback func(key string) (interface{}, error)

type IStore interface {
	OnEvicted(func(key string, obj []byte))
	Get(key string) (obj []byte, err error)
	GetExpired(key string) (obj []byte, expired time.Time, err error)
	Set(key string, obj []byte, dur ...time.Duration) error
	Delete(key string) error
	DeleteExpired() error
	Close() error
}
