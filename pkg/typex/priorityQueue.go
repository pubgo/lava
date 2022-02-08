package typex

import "container/heap"

type PriorityQueueItem struct {
	Value    interface{}
	Priority int64
	Index    int
}

// PriorityQueue implements the heap.Interface.
type PriorityQueue []*PriorityQueueItem

// Len returns the PriorityQueue length.
func (pq PriorityQueue) Len() int { return len(pq) }

// Less is the items less comparator.
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }

// Swap exchanges the indexes of the items.
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) PushItem(x *PriorityQueueItem) { pq.Push(x) }

// Push implements the heap.Interface.Push.
// Adds x as element Len().
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityQueueItem)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) PopItem() *PriorityQueueItem {
	var val = pq.Pop()
	if val == nil {
		return nil
	}

	return val.(*PriorityQueueItem)
}

// Pop implements the heap.Interface.Pop.
// Removes and returns element Len() - 1.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	if n == 0 {
		return nil
	}

	item := old[n-1]
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Head returns the first item of a PriorityQueue without removing it.
func (pq *PriorityQueue) Head() *PriorityQueueItem { return (*pq)[0] }

// Remove removes and returns the element at Index i from the PriorityQueue.
func (pq *PriorityQueue) Remove(i int) interface{} { return heap.Remove(pq, i) }
