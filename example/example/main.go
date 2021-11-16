package main

import (
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"io/ioutil"
)

func main() {
	var bytes, err = ioutil.ReadFile("./docs/swagger/proto/user/user.swagger.json")
	xerror.Panic(err)

	specDoc, err := loads.Analyzed(bytes, "")
	xerror.Panic(err)

	specDoc, err = specDoc.Expanded(&spec.ExpandOptions{
		SkipSchemas:         false,
		ContinueOnError:     true,
		AbsoluteCircularRef: true,
	})
	xerror.Panic(err)
	fmt.Println(specDoc.Spec().Host)
	for k,v:=range specDoc.Spec().Paths.Paths{
		fmt.Println(k,v)
		q.Q(v.Post)
	}
}
