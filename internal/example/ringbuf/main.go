package main

import "go.uber.org/atomic"

type Queue struct {
	cap      int32
	index    atomic.Int32
	data     [2000000]interface{}
	dataChan chan interface{}
}

func (t *Queue) loop() {
	go func() {
		for {
			select {
			case val := <-t.dataChan:
				if t.index.Load() == t.cap {
					t.index.Store(0)
				} else {
					t.index.Inc()
				}
				t.data[t.index.Load()] = val
			}
		}
	}()
}

func (t *Queue) Put(val interface{}) { t.dataChan <- val }

func (t *Queue) Range(fn func(val interface{}) bool) {
	var index = t.index.Load()

	for i := index; i >= 0; i-- {
		if !fn(t.data[i]) {
			return
		}
	}

	for i := t.cap - 1; i >= 0; i-- {
		if !fn(t.data[i]) {
			return
		}
	}
}

func main() {

}
