package p2c

import (
	"context"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
)

type mockSubConn struct {
}

func (msc *mockSubConn) Connect() {
}

func (msc *mockSubConn) UpdateAddresses([]resolver.Address) {
}

func TestP2cPick_Nil(t *testing.T) {
	p2b := new(p2cBalancer)
	picker := p2b.Build(nil)
	_, _, err := picker.Pick(context.Background(), balancer.PickInfo{
		FullMethodName: "/",
		Ctx:            context.Background(),
	})
	assert.NotNil(t, err)
}

func TestP2cPicker_Pick(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "single",
			count: 20,
		},
		{
			name:  "two",
			count: 1,
		},
		{
			name:  "multiple",
			count: 100,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			const total = 100000
			p2b := new(p2cBalancer)
			ready := make(map[resolver.Address]balancer.SubConn)
			for i := 0; i < test.count; i++ {
				ready[resolver.Address{
					Addr: strconv.Itoa(i),
				}] = new(mockSubConn)
			}

			p2cPk := p2b.Build(ready)
			var wg sync.WaitGroup
			wg.Add(total)
			for i := 0; i < total; i++ {
				_, done, err := p2cPk.Pick(context.Background(), balancer.PickInfo{
					FullMethodName: "/",
					Ctx:            context.Background(),
				})
				assert.Nil(t, err)
				if i%100 == 0 {
					err = status.Error(codes.DeadlineExceeded, "deadline")
				}
				go func() {
					runtime.Gosched()
					done(balancer.DoneInfo{
						Err: err,
					})
					wg.Done()
				}()
			}
		})
	}
}
