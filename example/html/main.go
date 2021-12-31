package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pubgo/lava/plugins/syncx"
	"github.com/pubgo/xerror"
)

func main() {
	var n = time.Now()
	defer func() {
		fmt.Println(time.Since(n))
	}()

	var wg syncx.WaitGroup
	for i := 0; i < 100; i++ {
		//wg.Go(func() {
		resp, err := http.Get("https://m.gongbocoins.com/k/1110000007")
		xerror.Panic(err)
		fmt.Println(resp.StatusCode)
		dd, err := ioutil.ReadAll(resp.Body)
		xerror.Panic(err)
		fmt.Println(string(dd))
		//})
	}
	wg.Wait()
}
