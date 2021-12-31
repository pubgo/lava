package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/bradleyjkemp/memviz"
	"github.com/pubgo/xerror"
)

type tree struct {
	id    int
	left  *tree
	right *tree
}

func main() {
	defer xerror.Resp(func(err xerror.XErr) {
		err.Debug()
	})

	var ss *tree
	fmt.Println(ss.id)

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

	buf := &bytes.Buffer{}
	memviz.Map(buf, &root)
	err := ioutil.WriteFile("example-tree-data", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}
