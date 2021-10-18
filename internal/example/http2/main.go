package main

import (
	"github.com/pubgo/xerror"
	"golang.org/x/net/http2"

	"fmt"
	"log"
	"net/http"
)

func main() {
	http2.VerboseLogs = true

	var srv http.Server
	srv.Addr = ":8080"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "HTTP协议: %s\n", r.Proto)
	})

	//h2c.NewHandler(handler, h2s)

	xerror.Panic(http2.ConfigureServer(&srv, &http2.Server{}))
	go func() {
		log.Fatal(srv.ListenAndServeTLS("./cert/server.crt", "./cert/server.key"))
	}()

	fmt.Println(srv.Addr)
	select {}
}
