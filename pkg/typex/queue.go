package typex

import (
	"sync"
)

func QueueOf(val ...interface{}) *Queue {
	var q = &Queue{}
	for i := range val {
		q.Push(val[i])
	}
	return q
}

type Queue struct {
	mu   sync.RWMutex
	data []interface{}
}

func (t *Queue) Push(val interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data = append(t.data, val)
}

func (t *Queue) Pop() interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	var data = make([]interface{}, len(t.data)-1)
	copy(data, t.data[:len(t.data)-2])
	t.data = data
	return t.data[len(t.data)-1]
}

func (t *Queue) PopFirst() interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.data)==0{
		return nil
	}

	var data = make([]interface{}, len(t.data)-1)
	copy(data, t.data[1:])
	t.data = data
	return t.data[0]
}

func (t *Queue) Del(index uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if uint32(len(t.data)) < index {
		return
	}

	copy(t.data[:index], t.data[index+1:])
}

func (t *Queue) Get(index uint32) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if uint32(len(t.data)) < index {
		return nil
	}

	return t.data[index]
}

func (t *Queue) List() []interface{} {
	var data = make([]interface{}, len(t.data))
	copy(data, t.data)
	return data
}

func (t *Queue) First() interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.data[0]
}

func (t *Queue) Last() interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.data[len(t.data)-1]
}
