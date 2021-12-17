package hello

import (
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name: "hello",
	Action: func(context *cli.Context) error {
		// Components routing:
		app.Route("/", &hello{})
		app.Route("/hello", &hello{})
		app.RunWhenOnBrowser()

		// HTTP routing:
		http.Handle("/", &app.Handler{
			Name:        "Hello",
			Description: "An Hello World! example",
		})

		http.ListenAndServe(":8088", nil)
		return nil
	},
}
