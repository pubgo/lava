package main

import (
	"fmt"
	"github.com/pubgo/xerror"
)

func main() {
	var data map[string]interface{}
	for k,v:=range data{
		_,_=k,v
	}
	var a, b = data["hello"]
	fmt.Println(a, b)

	defer xerror.RespExit()
	init1()
}

func init1() {
	defer xerror.Raise(init1)
	panic("ok")
}