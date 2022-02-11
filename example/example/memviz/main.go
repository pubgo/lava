package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pubgo/lava/pkg/utils"
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

	xerror.Panic(ioutil.WriteFile("example-tree-data", utils.Memviz(&root), 0644))
}
