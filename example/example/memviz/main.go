package main

import (
	"fmt"
	"github.com/pubgo/lava/pkg/reflectx"
	"io/ioutil"
	"reflect"

	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/xerror"
)

type tree struct {
	id    int
	left  *tree
	right *tree
}

type Abc interface {
}

type Abc1 = Abc

type AbcImpl struct {
	Abc1
}

func main() {
	defer xerror.Resp(func(err xerror.XErr) {
		err.Debug()
	})

	var ss = reflect.TypeOf((*Abc1)(nil)).Elem()

	var v1 = reflect.ValueOf(&AbcImpl{}).Elem()
	var field = reflectx.FindFieldBy(v1, func(field reflect.StructField) bool {
		return ss.String() == field.Type.String()
	})
	fmt.Println(field.Type().String())

	root := &tree{
		id: 0,
		left: &tree{
			id: 1,
		},
		right: &tree{
			id: 2,
		},
	}
	leaf := &tree{
		id: 3,
	}

	root.left.right = leaf
	root.right.left = leaf

	xerror.Panic(ioutil.WriteFile("example-tree-data", utils.Memviz(&root), 0644))
}
