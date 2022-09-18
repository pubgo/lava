package merge

import (
	"testing"

	"github.com/pubgo/x/q"
)

type dst struct {
	name  string
	Name  string
	Hello string `json:"hello"`
}

type src struct {
	Name  string `json:"name"`
	Hello string `json:"hello"`
}

func TestStruct(t *testing.T) {
	q.Q(Struct(&dst{name: "1", Hello: "1"}, src{Name: "2", Hello: "2"}))
	q.Q(Struct(&dst{name: "1", Hello: "1"}, &src{Name: "2", Hello: "2"}))

	var dd = &dst{name: "1", Hello: "1"}
	q.Q(Struct(&dd, &src{Name: "2", Hello: "2"}))

	var rr = &src{Name: "2", Hello: "2"}
	q.Q(Struct(&dd, &rr))

	var d1 = map[string]interface{}{"a": src{Name: "2", Hello: "2"}}
	var d2 = map[string]dst{"a": {Name: "1", Hello: "1"}, "b": {Name: "1", Hello: "1"}}
	Copy(&d1, &d2).Unwrap()
}

func TestMapStruct(t *testing.T) {
	q.Q(MapStruct(&dst{name: "1", Hello: "1"}, map[string]interface{}{"name": "2", "hello": "2"}))
	q.Q(MapStruct(&dst{name: "1", Hello: "1"}, &map[string]interface{}{"name": "2", "hello": "2"}))

	var dd map[string]dst
	q.Q(MapStruct(&dd, map[string]map[string]interface{}{"name": {"name": "2", "hello": "2"}, "hello": {"name": "2", "hello": "2"}}))

	//var rr = &map[string]interface{}{"name": "2", "hello": "2"}
	//q.Q(MapStruct(&dd, &rr)) // error
}
