package typex

import "sync"

type RwMap struct {
	rw   sync.RWMutex
	data map[string]interface{}
}

func (t *RwMap) Has(key string) bool {
	t.rw.RLock()
	defer t.rw.RUnlock()

	_, ok := t.data[key]
	return ok
}

func (t *RwMap) Map() map[string]interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()

	var dt = make(map[string]interface{}, len(t.data))
	for k, v := range t.data {
		dt[k] = v
	}

	return dt
}

func (t *RwMap) Get(key string) interface{} {
	t.rw.RLock()
	val, ok := t.data[key]
	defer t.rw.RUnlock()

	if ok {
		return val
	}

	return NotFound
}

func (t *RwMap) Load(key string) (interface{}, bool) {
	t.rw.RLock()
	val, ok := t.data[key]
	t.rw.RUnlock()

	return val, ok
}

func (t *RwMap) Keys() []string {
	t.rw.RLock()
	defer t.rw.RUnlock()

	var keys = make([]string, 0, len(t.data))
	for k := range t.data {
		keys = append(keys, k)
	}
	return keys
}

func (t *RwMap) Each(fn func(name string, val interface{})) {
	t.rw.RLock()
	defer t.rw.RUnlock()

	for k, v := range t.data {
		fn(k, v)
	}
}

func (t *RwMap) Range(fn func(name string, val interface{}) bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()

	for k, v := range t.data {
		if !fn(k, v) {
			break
		}
	}
}

func (t *RwMap) Set(key string, val interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()

	if t.data == nil {
		t.data = make(map[string]interface{}, 8)
	}

	t.data[key] = val
}

func (t *RwMap) Del(key string) {
	t.rw.Lock()
	defer t.rw.Unlock()

	delete(t.data, key)
}
