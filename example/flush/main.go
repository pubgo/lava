package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "x-request-id")

		// Listen to the closing of the http connection via the CloseNotifier
		notify := w.(http.CloseNotifier).CloseNotify()
		go func() {
			<-notify
			// Remove this client from the map of attached clients
			// when `EventHandler` exits.
			log.Println("HTTP connection just closed.")
		}()

		f, _ := w.(http.Flusher)

		for i := 0; i < 100; i++ {
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

	log.Fatal(http.ListenAndServe(":8888", nil))
}
