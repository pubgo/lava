package syncx

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func httpGetList() *Promise {
	return YieldGroup(func(in chan<- *Promise) error {
		for i := 2; i > 0; i-- {
			in <- Yield(func() (interface{}, error) { return http.Get("https://www.baidu.com") })
		}
		return nil
	})
}

func httpGet() *Promise {
	return Yield(func() (interface{}, error) {
		//time.After(time.Millisecond * 10)
		//panic("panic")
		return http.Get("https://www.baidu.com")
	})
}

func handleResp() (interface{}, error) {
	return httpGet().Await()
}

func TestPromise_Unwrap(t *testing.T) {
	GoDelay(func() {
		var p = httpGet()
		resp := <-p.Unwrap()
		fmt.Println("httpGet", p.Err(), resp)
	})

	GoDelay(func() {
		var out = httpGetList()
		for resp := range out.Unwrap() {
			fmt.Println("httpGetList", resp)
		}
		fmt.Println("httpGetList", out.Err())
	})
	<-time.After(time.Second)
}

func TestGoChan(t *testing.T) {
	var now = time.Now()
	defer func() {
		fmt.Println(time.Since(now))
	}()

	var val1 = GoChan(func() Value {
		time.Sleep(time.Millisecond)
		fmt.Println("2")
		//return WithErr(errors.New("error"))
		return WithVal("hello")
	})

	var val2 = GoChan(func() Value {
		time.Sleep(time.Millisecond)
		fmt.Println("3")
		//return WithErr(errors.New("error"))
		return WithVal("hello")
	})

	Wait(val1, val2)

	//_, _ = <-val1, <-val2

	fmt.Println("1", val2, val1)
}
