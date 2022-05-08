package inject

import (
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
	A *testObj
	B testInter
}

func TestRegister(t *testing.T) {
	Register((*testObj)(nil), func(obj Object, field Field) (interface{}, bool) {
		return &testObj{name: "hello struct"}, true
	})

	Register((*testInter)(nil), func(obj Object, field Field) (interface{}, bool) {
		return &testObj{name: "hello interface"}, true
	})

	var t1 = new(testInject)
	Inject(t1)
	if t1.A.name != "hello struct" {
		t.Fatal("inject failed")
	}

	if t1.B.Hello() != "hello interface" {
		t.Fatal("inject failed")
	}
}
