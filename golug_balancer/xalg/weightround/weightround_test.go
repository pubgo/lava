package weightround

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightRound(t *testing.T) {
	t.Run("no item", func(t *testing.T) {
		ll := NewWr()
		item, _ := ll.Next()
		assert.Nil(t, item)
	})

	t.Run("1 item", func(t *testing.T) {
		ll := NewWr()
		ll.Add("aa", 1, "a")
		item, _ := ll.Next()
		assert.Equal(t, "a", item)
	})

	t.Run("3 items", func(t *testing.T) {
		ll := NewWr()
		ll.Add("a", 2, "a")
		ll.Add("b", 3, "b")
		ll.Add("c", 5, "c")

		countMap := make(map[interface{}]int)

		totalCount := 10000
		for i := 0; i < totalCount; i++ {
			item, _ := ll.Next()
			countMap[item]++
		}

		assert.Equal(t, float64(countMap["a"])/float64(totalCount), 0.2)
		assert.Equal(t, float64(countMap["b"])/float64(totalCount), 0.3)
		assert.Equal(t, float64(countMap["c"])/float64(totalCount), 0.5)

		total := 0
		for _, count := range countMap {
			total += count
		}
		assert.Equal(t, totalCount, total)
	})

}
