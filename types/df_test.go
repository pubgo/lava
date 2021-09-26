package types

import (
	"fmt"
	"testing"
)

func TestName1(t *testing.T) {
	var df = &DataFrame{data: []interface{}{
		map[string]interface{}{
			"a":    "hello",
			"data": 1,
		},
		map[string]interface{}{
			"a":    "hello",
			"data": 1,
		},
		map[string]interface{}{
			"a":    "world",
			"data": 3,
		},
		map[string]interface{}{
			"a":    "world",
			"data": 4,
		},
	}}

	df = df.Filter("data<4")
	var agg = df.GroupBy("a")
	for k, v := range agg {
		fmt.Println(k, v.Sum("data"), v.Avg("data"))
	}
}
