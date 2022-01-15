package resource

import (
	"fmt"
	"io"
	"runtime"
	"testing"
)

type client struct {
	Resource
}

func (r *client) Kind() string { return "test-client" }
func (r *client) Get() (*res, func()) {
	var rr, release = r.Resource.Load()
	return rr.(*res), release
}

var _ io.Closer = (*res)(nil)

type res struct {
	name string
}

func (t *res) Close() error { return nil }

func TestName(t1 *testing.T) {
	func() {
		xx := &res{name: "123"}
		yy := &res{name: "456"}

		fmt.Printf("address for original %d, address for new %d\n", &xx, &yy)

		var dd = &client{New(xx)}
		Update("", dd)
		var rr, release = dd.Get()
		fmt.Println(rr.name)
		release()

		Update("", &client{New(yy)})
		rr, release = dd.Get()
		fmt.Println(rr.name)
		release()

		// 不会更新, yy对象未改变
		Update("", &client{New(yy)})
		rr, release = dd.Get()
		fmt.Println(rr.name)
		release()

		Remove("test-client", "default")
		dd = nil
	}()

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
}
