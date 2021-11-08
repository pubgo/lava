package plugin

import (
	"sync"
)

var beforeStarts []func()
var afterStarts []func()
var beforeStops []func()
var afterStops []func()
var mu sync.Mutex

func GetBeforeStartsList() []func() {
	mu.Lock()
	defer mu.Unlock()
	return beforeStarts
}

func GetAfterStartsList() []func() {
	mu.Lock()
	defer mu.Unlock()
	return afterStarts
}

func GetBeforeStopsList() []func() {
	mu.Lock()
	defer mu.Unlock()
	return beforeStops
}

func GetAfterStopsList() []func() {
	mu.Lock()
	defer mu.Unlock()
	return afterStops
}

func BeforeStart(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	beforeStarts = append(beforeStarts, fn)
}

func AfterStart(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	afterStarts = append(afterStarts, fn)
}

func BeforeStop(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	beforeStops = append(beforeStops, fn)
}

func AfterStop(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	afterStops = append(afterStops, fn)
}
