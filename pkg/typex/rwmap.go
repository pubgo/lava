package typex

import "sync"

type MapRWM struct {
	rw   sync.RWMutex
	data map[string]interface{}
}

func (t *MapRWM) Has(key string) bool {
	t.rw.RLock()
	defer t.rw.RUnlock()

	_, ok := t.data[key]
	return ok
}

func (t *MapRWM) Map() map[string]interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()

	var dt = make(map[string]interface{}, len(t.data))

	for k, v := range t.data {
		dt[k] = v
	}

	return dt
}

func (t *MapRWM) Get(key string) interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()

	val, ok := t.data[key]
	if ok {
		return val
	}

	return NotFound
}

func (t *MapRWM) Load(key string) (interface{}, bool) {
	t.rw.RLock()
	val, ok := t.data[key]
	t.rw.RUnlock()

	return val, ok
}

func (t *MapRWM) Keys() []string {
	t.rw.RLock()
	defer t.rw.RUnlock()

	var keys = make([]string, 0, len(t.data))
	for k := range t.data {
		keys = append(keys, k)
	}
	return keys
}

func (t *MapRWM) Each(fn func(name string, val interface{})) {
	t.rw.RLock()
	defer t.rw.RUnlock()

	for k, v := range t.data {
		fn(k, v)
	}
}

func (t *MapRWM) Set(key string, val interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()

	t.data[key] = &val
}

func (t *MapRWM) Del(key string) {
	t.rw.Lock()
	defer t.rw.Unlock()

	delete(t.data, key)
}
