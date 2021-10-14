package syncx

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func httpGet() *AsyncValue {
	return Async(func(ctx context.Context) (interface{}, error) {
		return http.Get("https://www.baidu.com")
	})
}

func TestGoChan(t *testing.T) {
	<-GoChan(func() {
		fmt.Println("2")
		panic("hello")
	}, func(err error) {
		//panic("hello")
	})

	fmt.Println("1")

	var resp = httpGet().Expect("http get error")
	fmt.Println(resp)
}
