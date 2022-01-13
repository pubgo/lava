package resource

import (
	"fmt"
	"io"
	"runtime"
	"testing"
)

var _ Resource = (*client)(nil)

type client struct {
	v *res
}

func (r *client) Unwrap() io.Closer      { return r.v }
func (r *client) UpdateObj(val Resource) { r.v = val.(*client).v }
func (r *client) Kind() string           { return "test-client" }
func (r *client) Get() *res              { return r.v }

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

		var dd = &client{v: xx}
		Update("", dd)
		fmt.Println(dd.Get().name)

		Update("", &client{v: yy})
		fmt.Println(dd.Get().name)

		// 不会更新, yy对象未改变
		Update("", &client{v: yy})
		fmt.Println(dd.Get().name)

		Remove("test-client", "default")
		dd = nil
	}()

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
}
