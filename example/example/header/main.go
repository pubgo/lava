package main

import (
	"github.com/kr/pretty"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

func main() {
	var h = fasthttp.RequestHeader{}
	h.Set("a", "b")
	h.Add("b", "c")
	h.Add("b", "d")
	var md = make(metadata.MD)
	h.VisitAll(func(key, value []byte) {
		md.Append(utils.BtoS(key), utils.BtoS(value))
	})
	pretty.Logln(md)
}
