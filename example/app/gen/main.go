package main

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pubgo/lug/example/app/hello"
	"github.com/pubgo/xerror"
)

func main() {
	// Components routing:
	app.Route("/", &hello.Hello{})
	app.Route("/hello", &hello.Hello{})

	app.RunWhenOnBrowser()

	xerror.Panic(app.GenerateStaticWebsite("./static", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Resources:   app.GitHubPages("REPOSITORY_NAME"),
	}))
}
