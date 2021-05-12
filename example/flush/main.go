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

		f, _ := w.(http.Flusher)

		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "time.now(): %v \n\r", time.Now())
			f.Flush()
			time.Sleep(time.Second)
		}

	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
