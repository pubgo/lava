package main

import (
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pubgo/lug/example/app/hello"
)

func main() {
	// Components routing:
	app.Route("/", &hello.Hello{})
	app.Route("/hello", &hello.Hello{})

	app.RunWhenOnBrowser()

	// HTTP routing:
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
	})

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
