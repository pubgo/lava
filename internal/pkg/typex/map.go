package typex

import "sync"

var mu sync.Mutex

type Map struct {
	done int8
	data map[string]interface{}
}

func (t *Map) once() {
	mu.Lock()
	defer mu.Unlock()

	t.done = 1
	t.data = make(map[string]interface{})
}

func (t *Map) check() {
	if t.done == 1 {
		return
	}

	t.once()
}

func (t *Map) Has(key string) bool {
	t.check()

	_, ok := t.data[key]
	return ok
}

func (t *Map) Map() map[string]interface{} {
	t.check()

	return t.data
}

func (t *Map) Get(key string) interface{} {
	t.check()

	val, ok := t.data[key]
	if ok {
		return val
	}

	return NotFound
}

func (t *Map) Load(key string) (interface{}, bool) {
	t.check()

	val, ok := t.data[key]
	return val, ok
}

func (t *Map) Keys() []string {
	t.check()

	var keys = make([]string, 0, len(t.data))
	for k := range t.data {
		keys = append(keys, k)
	}
	return keys
}

func (t *Map) Each(fn func(name string, val interface{})) {
	t.check()

	for k, v := range t.data {
		fn(k, v)
	}
}

func (t *Map) Set(key string, val interface{}) {
	t.check()

	t.data[key] = val
}

func (t *Map) Del(key string) {
	t.check()

	delete(t.data, key)
}
