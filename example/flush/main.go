package main

import (
	"fmt"
	"github.com/gin-contrib/sse"
	"github.com/pubgo/xerror"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "x-request-id")

		// Listen to the closing of the http connection via the CloseNotifier
		//notify := w.(http.CloseNotifier).CloseNotify()
		go func() {
			<-r.Context().Done()
			// Remove this client from the map of attached clients
			// when `EventHandler` exits.
			log.Println("HTTP connection just closed.")
		}()

		f, _ := w.(http.Flusher)



		for i := 0; i < 5; i++ {
			sse.Encode(w, sse.Event{
				Event: "message",
				Data:  "some data\nmore data",
			})
			sse.Encode(w, sse.Event{
				Id:    "124",
				Event: "message",
				Data: map[string]interface{}{
					"user":    "manu",
					"date":    time.Now().Unix(),
					"content": "hi!",
				},
			})
			fmt.Fprintf(w, "time.now(): %v \n\r", time.Now())
			f.Flush()
			time.Sleep(time.Second)
		}
	})

	// Set the headers related to event streaming.
	//w.Header().Set("Content-Type", "text/event-stream")
	//w.Header().Set("Cache-Control", "no-cache")
	//w.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Transfer-Encoding", "chunked")
	//fmt.Fprintf(w, "data: Message: %s\n\n", msg)

	go func() {
		time.Sleep(time.Second)
		//var cc, _ = restc.DefaultCfg().Build()
		//var resp, err = cc.Get("http://localhost:8888/")
		var resp, err = http.Get("http://localhost:8888/")
		xerror.Panic(err)
		//for {
		var ddd, err1 = ioutil.ReadAll(resp.Body)
		fmt.Println(string(ddd), err1)
		time.Sleep(time.Second)
		//}
	}()

	log.Fatal(http.ListenAndServe(":8888", nil))
}
