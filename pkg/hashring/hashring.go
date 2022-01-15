package hashring

import (
	"crypto/sha1"
	"math"
	"sort"
	"strconv"
	"sync"
)

const (
	defaultSpots = 10
)

type hashringNode struct {
	nodeKey   string
	spotValue uint32
}
type hashringNodes []hashringNode

func (p hashringNodes) Len() int           { return len(p) }
func (p hashringNodes) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p hashringNodes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p hashringNodes) Sort()              { sort.Sort(p) }

type HashRing struct {
	mutex   sync.RWMutex
	spots   int
	nodes   hashringNodes
	weights map[string]int
}

func NewHashRing(weights map[string]int, spotsArgs ...int) *HashRing {
	spots := defaultSpots
	if len(spotsArgs) > 0 && spotsArgs[0] > 0 {
		spots = spotsArgs[0]
	}

	h := &HashRing{
		spots:   spots,
		weights: weights,
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.generate()
}

func genValue(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	return v
}
func (h *HashRing) generate() *HashRing {
	var totalW int
	for _, w := range h.weights {
		totalW += w
	}

	totalVirtualSpots := h.spots * len(h.weights)
	h.nodes = hashringNodes{}

	for nodeKey, w := range h.weights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			hash := sha1.New()
			hash.Write([]byte(nodeKey + ":" + strconv.Itoa(i)))
			hashBytes := hash.Sum(nil)
			n := hashringNode{
				nodeKey:   nodeKey,
				spotValue: genValue(hashBytes[6:10]),
			}
			h.nodes = append(h.nodes, n)
			hash.Reset()
		}
	}
	h.nodes.Sort()

	return h
}

func (h *HashRing) Append(nodeKey string, weight int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.weights[nodeKey] = weight
	h.generate()
}

func (h *HashRing) Remove(nodeKey string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.weights, nodeKey)
	h.generate()
}

func (h *HashRing) Update(nodeKey string, weight int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.weights[nodeKey] = weight
	h.generate()
}

func (h *HashRing) Locate(s string) string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.nodes) == 0 {
		return ""
	}

	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	v := genValue(hashBytes[6:10])
	i := sort.Search(len(h.nodes), func(i int) bool { return h.nodes[i].spotValue >= v })

	if i == len(h.nodes) {
		i = 0
	}

	return h.nodes[i].nodeKey
}
