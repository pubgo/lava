package types

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	sm := NewSyncMap()
	sm.Set("a1", 1)
	sm.Set("a2", 2)
	fmt.Println(sm.Has("a1"))

	sm.Each(func(key string) {
		fmt.Println(key)
	})

	sm.Each(func(key string, val int) {
		fmt.Println(key, val)
	})

	var data = make(map[string]int)
	sm.Map(data)
	fmt.Println(data)

	var data1 map[string]int
	sm.Map(&data1)
	fmt.Println(data1)
}
