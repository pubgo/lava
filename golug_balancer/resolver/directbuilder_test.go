package resolver

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/grpc/resolver"
	"gotest.tools/assert"
)

func TestDirectBuilder_Scheme(t *testing.T) {
	var b directBuilder
	b.name = DirectScheme
	assert.Equal(t, DirectScheme, b.name)
}

func TestDirectBuilder_Build(t *testing.T) {
	nums := []int{0, 1, 2, subsetSize, subsetSize / 2, subsetSize * 2}
	for _, num := range nums {
		t.Run(strconv.Itoa(num), func(t *testing.T) {
			var servers []string
			for i := 0; i < num; i++ {
				servers = append(servers, fmt.Sprintf("localhost:%d", i))
			}
			var b directBuilder
			cc := new(mockedClientConn)
			_, err := b.Build(resolver.Target{
				Scheme:   DirectScheme,
				Endpoint: strings.Join(servers, ","),
			}, cc, resolver.BuildOptions{})
			assert.NilError(t, err)
			size := min(num, subsetSize)
			assert.Equal(t, size, len(cc.state.Addresses))
			m := make(map[string]struct{})
			for _, each := range cc.state.Addresses {
				m[each.Addr] = struct{}{}
			}
			assert.Equal(t, size, len(m))
		})
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}
