package cache

import "time"

var _ Store = (noopStore)(nil)

type noopStore struct{}

func (n noopStore) Get(key string) (interface{}, error)                        { return nil, nil }
func (n noopStore) Set(key string, val interface{}, ds ...time.Duration) error { return nil }

func (n noopStore) GetSet(key string, d time.Duration, fns ...getCallback) (interface{}, error) {
	return nil, nil
}

func (n noopStore) GetExpired(key string) (val interface{}, expireAt int64, err error) {
	return nil, 0, err
}

func (n noopStore) Delete(key string) error                       { return nil }
func (n noopStore) DeleteExpired() error                          { return nil }
func (n noopStore) OnEvicted(f func(key string, val interface{})) { return }
func (n noopStore) Close() error                                  { return nil }
