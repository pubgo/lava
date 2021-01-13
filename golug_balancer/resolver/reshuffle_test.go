package resolver

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubSet(t *testing.T) {
	tests := []struct {
		name  string
		set   int
		limit int
	}{
		{
			name:  "more",
			set:   100,
			limit: 36,
		},
		{
			name:  "less",
			set:   100,
			limit: 200,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var address []string
			for i := 0; i < test.set; i++ {
				address = append(address, strconv.Itoa(i))
			}

			set := reshuffle(address, test.limit)
			if test.limit < test.set {
				assert.Equal(t, test.limit, len(set))
			} else {
				assert.Equal(t, test.set, len(set))
			}
		})
	}
}
