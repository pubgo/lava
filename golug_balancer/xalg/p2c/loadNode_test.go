package p2c

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
)

func TestLoadNode(t *testing.T) {
	t.Run("0 item", func(t *testing.T) {
		ll := NewP2cAgl()
		item, _ := ll.Next()
		assert.Nil(t, item)
	})

	t.Run("1 item", func(t *testing.T) {
		ll := NewP2cAgl()
		ll.Add(1)
		item, _ := ll.Next()
		assert.Equal(t, 1, item)
	})

	t.Run("3 items", func(t *testing.T) {
		ll := NewP2cAgl()
		ll.Add(1)
		ll.Add(2)
		ll.Add(3)

		countMap := make(map[interface{}]int)

		totalCount := 10000
		for i := 0; i < totalCount; i++ {
			item, _ := ll.Next()
			countMap[item]++
		}

		total := 0
		for _, count := range countMap {
			total += count
			assert.Less(t, totalCount/3-200, count)
		}

		assert.Equal(t, totalCount, total)
	})
}

func TestLoadNodeAbnormal(t *testing.T) {
	t.Run("fixed inflight", func(t *testing.T) {
		ll := NewP2cAgl()
		ll.Add(1)
		ll.Add(2)
		ll.Add(3)

		item, _ := ll.Next()

		for i := 0; i < 1000; i++ {
			next, done := ll.Next()
			done(balancer.DoneInfo{})
			assert.NotEqual(t, item, next)
		}
	})
}
