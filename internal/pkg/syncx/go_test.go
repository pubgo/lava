package syncx

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pubgo/lava/internal/pkg/result"
)

func httpGetList() result.Chan[*http.Response] {
	return FromFuture(func(in chan<- *Future[*http.Response]) {
		for i := 2; i > 0; i-- {
			in <- Async(func() result.Result[*http.Response] {
				return result.New(http.Get("https://www.baidu.com"))
			})
		}
	})
}

func httpGet() *Future[*http.Response] {
	return Async(func() result.Result[*http.Response] {
		return result.New(http.Get("https://www.baidu.com"))
	})
}

func handleResp() result.Result[*http.Response] {
	return httpGet().Await()
}

func TestPromise_Unwrap(t *testing.T) {
	GoDelay(func() {
		var p = httpGet()
		resp := p.Await()
		fmt.Println("httpGet", resp.Err(), resp.Value())
	})

	GoDelay(func() {
		var out = httpGetList()
		for resp := range out {
			fmt.Println("httpGetList", resp.Err(), resp.Value())
		}
	})

	GoDelay(func() {
		var out = httpGetList()
		for resp := range out {
			fmt.Println("httpGetList", resp.Err(), resp.Value())
		}
	})
	<-time.After(time.Second)
}

func TestGoChan(t *testing.T) {
	var now = time.Now()
	defer func() {
		fmt.Println(time.Since(now))
	}()

	var val1 = Async(func() result.Result[string] {
		time.Sleep(time.Millisecond)
		fmt.Println("2")
		//return WithErr(errors.New("error"))
		return result.OK("hello")
	})

	var val2 = GoChan(func() result.Result[string] {
		time.Sleep(time.Millisecond)
		fmt.Println("3")
		//return WithErr(errors.New("error"))
		return result.OK("hello")
	})

	fmt.Println(Wait(val1, val2).ToResult().Value())
}
