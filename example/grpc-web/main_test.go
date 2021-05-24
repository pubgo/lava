package main

import (
	"fmt"
	"github.com/pubgo/xerror"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	//resp, err := http.Post("http://localhost:8900/hello.Transport/TestStream2", "application/grpc+json", strings.NewReader(`{"header":{"hello":"ok"}}`))
	resp, err := http.Get("http://localhost:8888")
	xerror.Panic(err)
	fmt.Println(resp.ContentLength)

	io.Copy(os.Stdout, resp.Body)
	//fmt.Println(resp.Body)
}
