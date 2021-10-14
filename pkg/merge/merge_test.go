package merge

import (
	"github.com/pubgo/x/q"
	"testing"
)

type dst struct {
	name  string
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
}

func TestMapStruct(t *testing.T) {
	q.Q(MapStruct(&dst{name: "1", Hello: "1"}, map[string]interface{}{"name": "2", "hello": "2"}))
	q.Q(MapStruct(&dst{name: "1", Hello: "1"}, &map[string]interface{}{"name": "2", "hello": "2"}))

	var dd = &dst{name: "1", Hello: "1"}
	q.Q(MapStruct(&dd, &map[string]interface{}{"name": "2", "hello": "2"}))

	//var rr = &map[string]interface{}{"name": "2", "hello": "2"}
	//q.Q(MapStruct(&dd, &rr)) // error
}
