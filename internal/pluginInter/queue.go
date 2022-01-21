package pluginInter

type item struct {
	Name     string
	Priority uint
}

type priorityQueue []item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(item))
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)

	if n == 0 {
		return nil
	}

	val := old[n-1]
	*pq = old[0 : n-1]
	return val
}
