package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pubgo/xerror"
)

func main() {
	var val = make(map[string]interface{})
	val["ssm"] = 2

	xerror.Panic(json.Unmarshal([]byte(`{"ss":1}`), &val))
	fmt.Printf("%#v\n\n", val)
	//xerror.Panic(json.Unmarshal([]byte(`[{"ss":1}]`), &val))
	//fmt.Printf("%#v\n\n", val)
	xerror.Panic(json.Unmarshal([]byte(`{"ss2":2,"ss1":2,"ss":3}`), &val))
	fmt.Printf("%#v\n\n", val)

	xerror.PanicErr(url.Parse("tcp4://:8080"))

	var a = []int{1, 2, 3, 4, 5, 6}
	var b = a[:0]
	fmt.Println(a, b)
	for i := 0; i < 40; i++ {
		b = append(b, i+100)
	}
	fmt.Println(a, b)
	for i := 0; i < len(a); i++ {
		a[i] = -1
	}
	fmt.Println(a, b)
}

func init1(a ...int) {

}
