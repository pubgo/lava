package typex

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestPQ(t *testing.T) {
	var pluginKeys PriorityQueue
	heap.Push(&pluginKeys, &PriorityQueueItem{Priority: 2, Value: 2})
	heap.Push(&pluginKeys, &PriorityQueueItem{Priority: 1, Value: 1})
	heap.Push(&pluginKeys, &PriorityQueueItem{Priority: 3, Value: 3})
	for {
		var item = pluginKeys.PopItem()
		if item == nil {
			break
		}

		fmt.Println(item.Value)
	}
}
