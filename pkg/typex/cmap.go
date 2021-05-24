package typex

import (
	"sync"
)

type CMap struct {
	rw   sync.RWMutex
	data map[string]*interface{}
}

func (t *CMap) Has(key string) bool {
	t.rw.RLock()
	_, ok := t.data[key]
	t.rw.RUnlock()

	return ok
}

func (t *CMap) Get(key string) interface{} {
	t.rw.RLock()
	val, ok := t.data[key]
	t.rw.RUnlock()

	if ok {
		return *val
	}

	return NotFound
}

func (t *CMap) GetSet(key string, val interface{}) (interface{}, bool) {
	t.rw.RLock()
	v, ok := t.data[key]
	t.rw.RUnlock()

	if ok {
		return *v, true
	}

	t.rw.Lock()
	v, ok = t.data[key]
	if ok {
		t.rw.Unlock()
		return *v, true
	}

	t.data[key] = &val
	t.rw.Unlock()

	return val, false
}

func (t *CMap) Keys() []string {
	t.rw.RLock()
	defer t.rw.RUnlock()

	var keys = make([]string, 0, len(t.data))
	for k := range t.data {
		keys = append(keys, k)
	}
	return keys
}

func (t *CMap) Set(key string, val interface{}) {
	t.rw.RLock()
	var v, ok = t.data[key]
	t.rw.RUnlock()

	if ok {
		*v = val
		return
	}

	t.rw.Lock()
	v, ok = t.data[key]
	if ok {
		*v = val
		t.rw.Unlock()
		return
	}

	t.data[key] = &val
	t.rw.Unlock()
}

func (t *CMap) Del(key string) {
	t.rw.RLock()
	var _, ok = t.data[key]
	t.rw.RUnlock()

	if !ok {
		return
	}

	t.rw.Lock()
	_, ok = t.data[key]
	if !ok {
		return
	}

	delete(t.data, key)
	t.rw.Unlock()
}
