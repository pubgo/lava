package main

import (
	"encoding/json"
	"fmt"

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
}
