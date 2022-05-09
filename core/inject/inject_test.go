package inject

import (
	"fmt"
	"testing"
)

type testInter interface {
	Hello() string
}

type testObj struct {
	name string
}

func (t *testObj) Hello() string { return t.name }

type testInject struct {
	Name string
	A    *testObj `inject:"name=${.Name}"`
	B    testInter
}

func TestRegister(t *testing.T) {
	Register((*testObj)(nil), func(obj Object, field Field) (interface{}, bool) {
		fmt.Println("testObj", field.Name())
		return &testObj{name: "hello struct"}, true
	})

	Register((*testInter)(nil), func(obj Object, field Field) (interface{}, bool) {
		fmt.Println("testInter", field.Name())
		return &testObj{name: "hello interface"}, true
	})

	var t1 = Inject(&testInject{Name: "jjjj"}).(*testInject)
	if t1.A.name != "hello struct" {
		t.Fatal("inject failed")
	}

	if t1.B.Hello() != "hello interface" {
		t.Fatal("inject failed")
	}
}
